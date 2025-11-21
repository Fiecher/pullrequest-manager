package pg

import (
	"context"
	"errors"
	"fmt"
	"pullrequest-manager/internal/domain/models"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrPullRequestNotFound = errors.New("pull request not found")

type PullRequestRepository struct {
	db *pgxpool.Pool
}

func NewPullRequestRepository(db *pgxpool.Pool) *PullRequestRepository {
	return &PullRequestRepository{db: db}
}

const (
	insertPullRequestQuery = `
		INSERT INTO pull_requests (title, author_id, status_id, merged_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at;
	`
	selectPullRequestByIDQuery = `
		SELECT id, title, author_id, status_id, merged_at, created_at, updated_at
		FROM pull_requests
		WHERE id = $1;
	`
	selectAllPullRequestsQuery = `
		SELECT id, title, author_id, status_id, merged_at, created_at, updated_at
		FROM pull_requests
		ORDER BY created_at DESC;
	`
	updatePullRequestQuery = `
		UPDATE pull_requests
		SET title = $1, author_id = $2, status_id = $3, merged_at = $4, updated_at = now()
		WHERE id = $5
		RETURNING updated_at;
	`
	deletePullRequestQuery = `
		DELETE FROM pull_requests WHERE id = $1;
	`
	selectByAuthorQuery = `
		SELECT id, title, author_id, status_id, merged_at, created_at, updated_at
		FROM pull_requests
		WHERE author_id = $1
		ORDER BY created_at DESC;
	`
	insertReviewerQuery = `
		INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id, assigned_at)
		VALUES ($1, $2, $3);
	`
	deleteReviewersQuery = `
		DELETE FROM pull_request_reviewers WHERE pull_request_id = $1;
	`
	selectReviewersQuery = `
		SELECT reviewer_id FROM pull_request_reviewers
		WHERE pull_request_id = $1;
	`
)

func (r *PullRequestRepository) Create(ctx context.Context, pr *models.PullRequest) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("create transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := tx.QueryRow(
		ctx,
		insertPullRequestQuery,
		pr.Title,
		pr.AuthorID,
		pr.StatusID,
		pr.MergedAt,
	).Scan(&pr.ID, &pr.CreatedAt, &pr.UpdatedAt); err != nil {
		return fmt.Errorf("insert pull request: %w", err)
	}

	if len(pr.ReviewersIDs) > 0 {
		if err := r.insertReviewersTx(ctx, tx, pr.ID, pr.ReviewersIDs); err != nil {
			return fmt.Errorf("insert reviewers: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (r *PullRequestRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.PullRequest, error) {
	var pr models.PullRequest

	err := r.db.QueryRow(
		ctx,
		selectPullRequestByIDQuery,
		id,
	).Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.StatusID, &pr.MergedAt, &pr.CreatedAt, &pr.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPullRequestNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find pull request by id %s: %w", id, err)
	}

	reviewers, err := r.getReviewers(ctx, pr.ID)
	if err != nil {
		return nil, fmt.Errorf("get reviewers for PR %s: %w", pr.ID, err)
	}
	pr.ReviewersIDs = reviewers

	return &pr, nil
}

func (r *PullRequestRepository) FindAll(ctx context.Context) ([]*models.PullRequest, error) {
	rows, err := r.db.Query(ctx, selectAllPullRequestsQuery)
	if err != nil {
		return nil, fmt.Errorf("find all pull requests: %w", err)
	}
	defer rows.Close()

	var list []*models.PullRequest

	for rows.Next() {
		var pr models.PullRequest
		if err := rows.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.StatusID, &pr.MergedAt, &pr.CreatedAt, &pr.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan pull request: %w", err)
		}

		reviewers, err := r.getReviewers(ctx, pr.ID)
		if err != nil {
			return nil, fmt.Errorf("get reviewers for PR %s: %w", pr.ID, err)
		}
		pr.ReviewersIDs = reviewers

		list = append(list, &pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating over pull request rows: %w", err)
	}

	return list, nil
}

func (r *PullRequestRepository) Update(ctx context.Context, pr *models.PullRequest) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("start transaction for update: %w", err)
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(
		ctx,
		updatePullRequestQuery,
		pr.Title,
		pr.AuthorID,
		pr.StatusID,
		pr.MergedAt,
		pr.ID,
	).Scan(&pr.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrPullRequestNotFound
	}
	if err != nil {
		return fmt.Errorf("update pull request %s: %w", pr.ID, err)
	}

	if _, err := tx.Exec(ctx, deleteReviewersQuery, pr.ID); err != nil {
		return fmt.Errorf("clear reviewers for PR %s: %w", pr.ID, err)
	}

	if len(pr.ReviewersIDs) > 0 {
		if err := r.insertReviewersTx(ctx, tx, pr.ID, pr.ReviewersIDs); err != nil {
			return fmt.Errorf("re-insert reviewers for PR %s: %w", pr.ID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction for update: %w", err)
	}

	return nil
}

func (r *PullRequestRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	cmd, err := r.db.Exec(ctx, deletePullRequestQuery, id)
	if err != nil {
		return fmt.Errorf("delete pull request %s: %w", id, err)
	}

	if cmd.RowsAffected() == 0 {
		return ErrPullRequestNotFound
	}

	return nil
}

func (r *PullRequestRepository) FindByAuthor(ctx context.Context, authorID uuid.UUID) ([]*models.PullRequest, error) {
	rows, err := r.db.Query(ctx, selectByAuthorQuery, authorID)
	if err != nil {
		return nil, fmt.Errorf("get pull requests by author %s: %w", authorID, err)
	}
	defer rows.Close()

	var list []*models.PullRequest

	for rows.Next() {
		var pr models.PullRequest
		if err := rows.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.StatusID, &pr.MergedAt, &pr.CreatedAt, &pr.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan pull request: %w", err)
		}

		reviewers, err := r.getReviewers(ctx, pr.ID)
		if err != nil {
			return nil, fmt.Errorf("get reviewers for PR %s: %w", pr.ID, err)
		}
		pr.ReviewersIDs = reviewers

		list = append(list, &pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating over pull request rows for author %s: %w", authorID, err)
	}

	return list, nil
}

func (r *PullRequestRepository) getReviewers(ctx context.Context, prID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.db.Query(ctx, selectReviewersQuery, prID)
	if err != nil {
		return nil, fmt.Errorf("get reviewers for PR %s: %w", prID, err)
	}
	defer rows.Close()

	var reviewers []uuid.UUID

	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan reviewer ID for PR %s: %w", prID, err)
		}
		reviewers = append(reviewers, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating over reviewer rows for PR %s: %w", prID, err)
	}

	return reviewers, nil
}

func (r *PullRequestRepository) insertReviewersTx(ctx context.Context, tx pgx.Tx, prID uuid.UUID, reviewers []uuid.UUID) error {
	for _, reviewer := range reviewers {
		_, err := tx.Exec(ctx, insertReviewerQuery, prID, reviewer, time.Now())
		if err != nil {
			return fmt.Errorf("insert reviewer %s for PR %s: %w", reviewer, prID, err)
		}
	}
	return nil
}
