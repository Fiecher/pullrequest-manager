package repositories

import (
	"pullrequest-inator/internal/infrastructure/models"

	"github.com/google/uuid"
)

type User interface {
	Repository[models.User, uuid.UUID]
}
