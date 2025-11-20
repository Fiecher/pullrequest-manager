package models

import (
	"github.com/google/uuid"
)

type Status struct {
	ID   uuid.UUID `db:"id"`
	Name string    `db:"name"`
}
