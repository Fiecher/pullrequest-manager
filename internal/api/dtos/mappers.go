package dtos

import (
	"pullrequest-inator/internal/infrastructure/models"

	"github.com/google/uuid"
)

func ModelToPullRequestDTO(pr *models.PullRequest, statusName string) *PullRequest {
	return &PullRequest{
		PullRequestId:     pr.ID.String(),
		PullRequestName:   pr.Title,
		AuthorId:          pr.AuthorID.String(),
		Status:            PullRequestStatus(statusName),
		AssignedReviewers: uuidsToStrings(pr.ReviewersIDs),
		CreatedAt:         &pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}

func uuidsToStrings(ids []uuid.UUID) []string {
	strings := make([]string, len(ids))
	for i, id := range ids {
		strings[i] = id.String()
	}
	return strings
}
