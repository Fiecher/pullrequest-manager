package repositories

import (
	"context"
	"pullrequest-inator/internal/domain/models"

	"github.com/google/uuid"
)

type Team interface {
	Repository[models.Team, uuid.UUID]
	FindByName(ctx context.Context, name string) (*models.Team, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Team, error)
}
