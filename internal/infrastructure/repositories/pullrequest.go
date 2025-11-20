package repositories

import (
	"pullrequest-manager/internal/domain/models"

	"github.com/google/uuid"
)

type PullRequest interface {
	Repository[models.PullRequest, uuid.UUID]
	FindByAuthor(userID uuid.UUID) ([]*models.PullRequest, error)
}
