package repositories

import (
	"pullrequest-manager/internal/domain/models"

	"github.com/google/uuid"
)

type Status interface {
	Repository[models.Status, uuid.UUID]
}
