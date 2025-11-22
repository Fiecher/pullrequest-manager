package repositories

import (
	"context"
	"pullrequest-inator/internal/api/dtos"
	"pullrequest-inator/internal/infrastructure/models"

	"github.com/google/uuid"
)

type Team interface {
	Repository[models.Team, uuid.UUID]
	FindByName(ctx context.Context, name string) (*models.Team, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Team, error)
	CreateWithUsers(ctx context.Context, teamReq *dtos.Team) error
}
