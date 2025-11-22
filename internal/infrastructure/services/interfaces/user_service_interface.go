package services

import (
	"context"
	"pullrequest-inator/internal/api/dtos"

	"github.com/google/uuid"
)

type User interface {
	RegisterUser(ctx context.Context, user *dtos.User) error
	UnregisterUserByID(ctx context.Context, userID uuid.UUID) error
	ListUsers(ctx context.Context) ([]*dtos.User, error)
	SetUserActive(ctx context.Context, userID uuid.UUID, active bool) (*dtos.User, error)
}
