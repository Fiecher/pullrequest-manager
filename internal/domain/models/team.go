package models

import (
	"time"

	"github.com/google/uuid"
)

type Team struct {
	ID        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	UserIDs []uuid.UUID
}
