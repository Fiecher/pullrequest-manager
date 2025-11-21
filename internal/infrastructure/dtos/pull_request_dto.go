package dtos

import (
	"time"

	"github.com/google/uuid"
)

type PullRequestDTO struct {
	PullRequestID     uuid.UUID   `json:"pull_request_id"`
	PullRequestName   string      `json:"pull_request_name"`
	AuthorID          uuid.UUID   `json:"author_id"`
	Status            string      `json:"status"`
	AssignedReviewers []uuid.UUID `json:"assigned_reviewers"`
	CreatedAt         *time.Time  `json:"createdAt,omitempty"`
	MergedAt          *time.Time  `json:"mergedAt,omitempty"`
}

type PullRequestShortDTO struct {
	PullRequestID   uuid.UUID `json:"pull_request_id"`
	PullRequestName string    `json:"pull_request_name"`
	AuthorID        uuid.UUID `json:"author_id"`
	Status          string    `json:"status"`
}

type ReassignReviewerResponseDTO struct {
	Pr         PullRequestDTO `json:"pr"`
	ReplacedBy uuid.UUID      `json:"replaced_by"`
}
