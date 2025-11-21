package repositories

import (
	"context"
	"pullrequest-inator/internal/domain/models"

	"github.com/google/uuid"
)

type PullRequest interface {
	Repository[models.PullRequest, uuid.UUID]
	FindByAuthor(ctx context.Context, userID uuid.UUID) ([]*models.PullRequest, error)
}
