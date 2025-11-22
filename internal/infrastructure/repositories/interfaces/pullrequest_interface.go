package repositories

import (
	"context"
	"pullrequest-inator/internal/infrastructure/models"

	"github.com/google/uuid"
)

type PullRequest interface {
	Repository[models.PullRequest, uuid.UUID]
	FindByReviewer(ctx context.Context, userID uuid.UUID) ([]*models.PullRequest, error)
}
