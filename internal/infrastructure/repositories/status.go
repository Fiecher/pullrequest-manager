package repositories

import (
	"context"
	"pullrequest-manager/internal/domain/models"

	"github.com/google/uuid"
)

type Status interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.Status, error)
	FindAll(ctx context.Context) ([]*models.Status, error)
}
