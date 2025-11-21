package repositories

import (
	"pullrequest-manager/internal/domain/models"

	"github.com/google/uuid"
)

type User interface {
	Repository[models.User, uuid.UUID]
}
