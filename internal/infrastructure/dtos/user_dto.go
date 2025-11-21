package dtos

import "github.com/google/uuid"

type UserDTO struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	TeamName string    `json:"team_name"`
	IsActive bool      `json:"is_active"`
}

type UserSetActiveRequestDTO struct {
	UserID   uuid.UUID `json:"user_id"`
	IsActive bool      `json:"is_active"`
}

type UserGetReviewResponseDTO struct {
	UserID       uuid.UUID             `json:"user_id"`
	PullRequests []PullRequestShortDTO `json:"pull_requests"`
}
