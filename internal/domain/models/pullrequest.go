package models

import (
	"time"

	"github.com/google/uuid"
)

type PullRequest struct {
	ID        uuid.UUID  `db:"id"`
	Title     string     `db:"title"`
	AuthorID  uuid.UUID  `db:"author_id"`
	StatusID  uuid.UUID  `db:"status_id"`
	MergedAt  *time.Time `db:"merged_at"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`

	ReviewersIDs []uuid.UUID
}
