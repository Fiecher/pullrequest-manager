package repositories

import (
	"pullrequest-manager/internal/domain/models"

	"github.com/google/uuid"
)

type Team interface {
	Repository[models.Team, uuid.UUID]
	FindByName(name string) (*models.Team, error)
	FindByUserID(userID uuid.UUID) (*models.Team, error)
}
