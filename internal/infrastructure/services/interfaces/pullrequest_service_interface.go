package services

import (
	"context"
	"pullrequest-inator/internal/api/dtos"

	"github.com/google/uuid"
)

type PullRequest interface {
	CreatePullRequest(ctx context.Context, pr *dtos.PullRequest) (*dtos.PullRequest, error)
	ReassignReviewer(ctx context.Context, userID uuid.UUID, prID uuid.UUID) (*dtos.ReassignReviewerResponse, error)
	FindPullRequestsByReviewer(ctx context.Context, userID uuid.UUID) ([]*dtos.PullRequest, error)
	MarkAsMerged(ctx context.Context, prID uuid.UUID) (*dtos.PullRequest, error)
	GetUserReviews(ctx context.Context, userID uuid.UUID) (*dtos.UserGetReviewResponse, error)
	CreateWithReviewers(ctx context.Context, prID uuid.UUID, prName string, authorID uuid.UUID) (*dtos.PullRequest, error)
}
