package pg

import (
	"context"
	"errors"
	"pullrequest-inator/internal/domain/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrStatusNotFound = errors.New("status not found")

type StatusRepository struct {
	db *pgxpool.Pool
}

func NewStatusRepository(db *pgxpool.Pool) *StatusRepository {
	return &StatusRepository{db: db}
}

const (
	getStatusByIDQuery = `SELECT id, name FROM pull_request_statuses WHERE id = $1`
	listStatusesQuery  = `SELECT id, name FROM pull_request_statuses ORDER BY name`
)

func (r *StatusRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Status, error) {
	var s models.Status
	if err := r.db.QueryRow(ctx, getStatusByIDQuery, id).Scan(&s.ID, &s.Name); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrStatusNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *StatusRepository) FindAll(ctx context.Context) ([]*models.Status, error) {
	rows, err := r.db.Query(ctx, listStatusesQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	statuses := make([]*models.Status, 0, 8)
	for rows.Next() {
		var s models.Status
		if err := rows.Scan(&s.ID, &s.Name); err != nil {
			return nil, err
		}
		statuses = append(statuses, &s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return statuses, nil
}
