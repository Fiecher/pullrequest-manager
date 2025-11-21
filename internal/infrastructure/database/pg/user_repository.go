package pg

import (
	"context"
	"errors"
	"fmt"
	"pullrequest-manager/internal/domain/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

const (
	insertUserQuery     = `INSERT INTO users (username, is_active) VALUES ($1, $2) RETURNING id, created_at, updated_at;`
	selectUserByIDQuery = `SELECT id, username, is_active, created_at, updated_at FROM users WHERE id = $1;`
	selectAllUsersQuery = `SELECT id, username, is_active, created_at, updated_at FROM users ORDER BY created_at DESC;`
	updateUserQuery     = `UPDATE users SET username = $1, is_active = $2, updated_at = now() WHERE id = $3 RETURNING updated_at;`
	deleteUserQuery     = `DELETE FROM users WHERE id = $1;`
)

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	if err := r.db.QueryRow(ctx, insertUserQuery, user.Username, user.IsActive).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	u := models.User{}

	err := r.db.QueryRow(ctx, selectUserByIDQuery, id).
		Scan(&u.ID, &u.Username, &u.IsActive, &u.CreatedAt, &u.UpdatedAt) // Добавлены поля времени, если они есть

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user by id %s: %w", id, err)
	}

	return &u, nil
}

func (r *UserRepository) FindAll(ctx context.Context) ([]*models.User, error) {
	rows, err := r.db.Query(ctx, selectAllUsersQuery)
	if err != nil {
		return nil, fmt.Errorf("find all users: %w", err)
	}
	defer rows.Close()

	var list []*models.User

	for rows.Next() {
		var u models.User
		if err = rows.Scan(&u.ID, &u.Username, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		list = append(list, &u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating over user rows: %w", err)
	}

	return list, nil
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	err := r.db.QueryRow(
		ctx,
		updateUserQuery,
		user.Username,
		user.IsActive,
		user.ID,
	).Scan(&user.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrUserNotFound
	}
	if err != nil {
		return fmt.Errorf("update user %s: %w", user.ID, err)
	}

	return nil
}

func (r *UserRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	cmd, err := r.db.Exec(ctx, deleteUserQuery, id)
	if err != nil {
		return fmt.Errorf("delete user %s: %w", id, err)
	}

	if cmd.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}
