package services

import (
	"context"
	"pullrequest-inator/internal/api/dtos"

	"github.com/google/uuid"
)

type Team interface {
	CreateTeamWithUsers(ctx context.Context, teamReq *dtos.Team) error
	GetTeamByName(ctx context.Context, teamName string) (*dtos.Team, error)
	SetUserActiveByID(ctx context.Context, userID uuid.UUID, active bool) (*dtos.User, error)
}
