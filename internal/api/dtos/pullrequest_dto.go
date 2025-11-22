package dtos

import "time"

type PullRequest struct {
	AssignedReviewers []string          `json:"assigned_reviewers"`
	AuthorId          string            `json:"author_id"`
	CreatedAt         *time.Time        `json:"createdAt"`
	MergedAt          *time.Time        `json:"mergedAt"`
	PullRequestId     string            `json:"pull_request_id"`
	PullRequestName   string            `json:"pull_request_name"`
	Status            PullRequestStatus `json:"status"`
}

type PullRequestStatus string

type PullRequestShort struct {
	AuthorId        string            `json:"author_id"`
	PullRequestId   string            `json:"pull_request_id"`
	PullRequestName string            `json:"pull_request_name"`
	Status          PullRequestStatus `json:"status"`
}
