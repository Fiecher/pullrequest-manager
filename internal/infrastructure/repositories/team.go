package repositories

import (
	"context"
	"pullrequest-manager/internal/domain/models"

	"github.com/google/uuid"
)

type Team interface {
	Repository[models.Team, uuid.UUID]
	FindByName(ctx context.Context, name string) (*models.Team, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Team, error)
}
