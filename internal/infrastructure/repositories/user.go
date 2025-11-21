package repositories

import (
	"pullrequest-inator/internal/domain/models"

	"github.com/google/uuid"
)

type User interface {
	Repository[models.User, uuid.UUID]
}
