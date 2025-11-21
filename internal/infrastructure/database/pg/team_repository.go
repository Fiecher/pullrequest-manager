package pg

import (
	"context"
	"errors"
	"pullrequest-inator/internal/domain/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrTeamNotFound = errors.New("team not found")

type TeamRepository struct {
	db *pgxpool.Pool
}

func NewTeamRepository(db *pgxpool.Pool) *TeamRepository {
	return &TeamRepository{db: db}
}

const (
	insertTeamQuery         = `INSERT INTO teams (name) VALUES ($1) RETURNING id`
	updateTeamQuery         = `UPDATE teams SET name=$1 WHERE id=$2`
	deleteTeamUsersQuery    = `DELETE FROM team_user WHERE team_id=$1`
	deleteTeamQuery         = `DELETE FROM teams WHERE id=$1`
	insertTeamUserQuery     = `INSERT INTO team_user (team_id, user_id) VALUES ($1, $2)`
	selectTeamByIDQuery     = `SELECT id, name, created_at, updated_at FROM teams WHERE id=$1`
	selectTeamByNameQuery   = `SELECT id, name, created_at, updated_at FROM teams WHERE name=$1`
	selectTeamUsersQuery    = `SELECT user_id FROM team_user WHERE team_id=$1`
	selectAllTeamsQuery     = `SELECT id, name, created_at, updated_at FROM teams ORDER BY created_at DESC`
	selectTeamByUserIDQuery = `SELECT t.id, t.name, t.created_at, t.updated_at FROM teams t JOIN team_user tu ON t.id = tu.team_id WHERE tu.user_id = $1`
)

func (r *TeamRepository) Create(ctx context.Context, team *models.Team) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := tx.QueryRow(ctx, insertTeamQuery, team.Name).Scan(&team.ID); err != nil {
		return err
	}

	for _, uid := range team.UserIDs {
		if _, err := tx.Exec(ctx, insertTeamUserQuery, team.ID, uid); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *TeamRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Team, error) {
	team := &models.Team{UserIDs: []uuid.UUID{}}

	err := r.db.QueryRow(ctx, selectTeamByIDQuery, id).Scan(
		&team.ID, &team.Name, &team.CreatedAt, &team.UpdatedAt,
	)
	if err != nil {
		return nil, ErrTeamNotFound
	}

	rows, err := r.db.Query(ctx, selectTeamUsersQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var uid uuid.UUID
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		team.UserIDs = append(team.UserIDs, uid)
	}

	return team, nil
}

func (r *TeamRepository) FindAll(ctx context.Context) ([]*models.Team, error) {
	rows, err := r.db.Query(ctx, selectAllTeamsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams := make([]*models.Team, 0)
	for rows.Next() {
		var t models.Team
		t.UserIDs = []uuid.UUID{}

		if err := rows.Scan(&t.ID, &t.Name, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}

		memberRows, err := r.db.Query(ctx, selectTeamUsersQuery, t.ID)
		if err != nil {
			return nil, err
		}
		for memberRows.Next() {
			var uid uuid.UUID
			if err := memberRows.Scan(&uid); err != nil {
				memberRows.Close()
				return nil, err
			}
			t.UserIDs = append(t.UserIDs, uid)
		}
		memberRows.Close()
		teams = append(teams, &t)
	}

	return teams, rows.Err()
}

func (r *TeamRepository) Update(ctx context.Context, team *models.Team) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	cmd, err := tx.Exec(ctx, updateTeamQuery, team.Name, team.ID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrTeamNotFound
	}

	if _, err = tx.Exec(ctx, deleteTeamUsersQuery, team.ID); err != nil {
		return err
	}

	for _, uid := range team.UserIDs {
		if _, err := tx.Exec(ctx, insertTeamUserQuery, team.ID, uid); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *TeamRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	cmd, err := tx.Exec(ctx, deleteTeamQuery, id)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrTeamNotFound
	}

	return tx.Commit(ctx)
}

func (r *TeamRepository) FindByName(ctx context.Context, name string) (*models.Team, error) {
	team := &models.Team{UserIDs: []uuid.UUID{}}

	err := r.db.QueryRow(ctx, selectTeamByNameQuery, name).Scan(
		&team.ID, &team.Name, &team.CreatedAt, &team.UpdatedAt,
	)
	if err != nil {
		return nil, ErrTeamNotFound
	}

	rows, err := r.db.Query(ctx, selectTeamUsersQuery, team.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var uid uuid.UUID
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		team.UserIDs = append(team.UserIDs, uid)
	}

	return team, nil
}

func (r *TeamRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Team, error) {
	team := &models.Team{UserIDs: []uuid.UUID{}}

	err := r.db.QueryRow(ctx, selectTeamByUserIDQuery, userID).Scan(
		&team.ID, &team.Name, &team.CreatedAt, &team.UpdatedAt,
	)
	if err != nil {
		return nil, ErrTeamNotFound
	}

	rows, err := r.db.Query(ctx, selectTeamUsersQuery, team.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var uid uuid.UUID
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		team.UserIDs = append(team.UserIDs, uid)
	}

	return team, nil
}
