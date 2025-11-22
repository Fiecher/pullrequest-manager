package services

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"pullrequest-inator/internal/api/dtos"
	"pullrequest-inator/internal/infrastructure/models"
	"pullrequest-inator/internal/infrastructure/repositories/interfaces"
	"pullrequest-inator/internal/infrastructure/repositories/pg"
	"time"

	"github.com/google/uuid"
)

var (
	ErrAuthorNotFound     = errors.New("author not found")
	ErrTeamNotFound       = errors.New("team not found")
	ErrPRAlreadyExists    = errors.New("pull request already exists")
	ErrPRNotFound         = errors.New("pull request not found")
	ErrUserNotReviewer    = errors.New("user is not a reviewer")
	ErrNoReviewCandidates = errors.New("no users available to review")
	ErrPRAlreadyMerged    = errors.New("cannot change PR state because already merged")
)

type PullRequestService struct {
	userRepo   repositories.User
	prRepo     repositories.PullRequest
	teamRepo   repositories.Team
	statusRepo repositories.Status
}

func NewPullRequestService(userRepo repositories.User, prRepo repositories.PullRequest,
	teamRepo repositories.Team, statusRepo repositories.Status) (*PullRequestService, error) {
	return &PullRequestService{
		userRepo:   userRepo,
		prRepo:     prRepo,
		teamRepo:   teamRepo,
		statusRepo: statusRepo,
	}, nil
}

func (s *PullRequestService) CreateWithReviewers(ctx context.Context, prID uuid.UUID,
	prName string, authorID uuid.UUID) (*dtos.PullRequest, error) {
	existing, err := s.prRepo.FindByID(ctx, prID)
	if err != nil && !errors.Is(err, pg.ErrPullRequestNotFound) {
		return nil, fmt.Errorf("check for existing PR: %w", err)
	}
	if existing != nil {
		return nil, ErrPRAlreadyExists
	}

	if _, err := s.userRepo.FindByID(ctx, authorID); errors.Is(err, pg.ErrUserNotFound) {
		return nil, ErrAuthorNotFound
	} else if err != nil {
		return nil, fmt.Errorf("find author: %w", err)
	}

	team, err := s.teamRepo.FindByUserID(ctx, authorID)
	if errors.Is(err, pg.ErrTeamNotFound) {
		return nil, ErrTeamNotFound
	} else if err != nil {
		return nil, fmt.Errorf("find team for author: %w", err)
	}

	var activeUsers []uuid.UUID
	for _, uid := range team.UserIDs {
		if uid == authorID {
			continue
		}
		u, err := s.userRepo.FindByID(ctx, uid)
		if errors.Is(err, pg.ErrUserNotFound) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("find user %s: %w", uid, err)
		}
		if u.IsActive {
			activeUsers = append(activeUsers, uid)
		}
	}

	if len(activeUsers) == 0 {
		return nil, ErrNoReviewCandidates
	}

	reviewers := chooseRandomUsers(activeUsers, 2)

	openStatus, err := s.getOpenStatus(ctx)
	if err != nil {
		return nil, err
	}

	newPR := &models.PullRequest{
		ID:           prID,
		Title:        prName,
		AuthorID:     authorID,
		StatusID:     openStatus.ID,
		MergedAt:     nil,
		ReviewersIDs: reviewers,
	}

	if err := s.prRepo.Create(ctx, newPR); err != nil {
		return nil, fmt.Errorf("create pull request: %w", err)
	}

	return dtos.ModelToPullRequestDTO(newPR, openStatus.Name), nil
}

func (s *PullRequestService) ReassignReviewer(ctx context.Context, userID uuid.UUID, prID uuid.UUID) (*dtos.ReassignReviewerResponse, error) {
	pr, err := s.prRepo.FindByID(ctx, prID)
	if errors.Is(err, pg.ErrPullRequestNotFound) {
		return nil, ErrPRNotFound
	} else if err != nil {
		return nil, fmt.Errorf("find PR: %w", err)
	}

	status, err := s.statusRepo.FindByID(ctx, pr.StatusID)
	if err != nil && !errors.Is(err, pg.ErrStatusNotFound) {
		return nil, fmt.Errorf("find status: %w", err)
	}
	if status != nil && status.Name == "MERGED" {
		return nil, ErrPRAlreadyMerged
	}

	reviewerIndex := -1
	for i, rid := range pr.ReviewersIDs {
		if rid == userID {
			reviewerIndex = i
			break
		}
	}
	if reviewerIndex == -1 {
		return nil, ErrUserNotReviewer
	}

	team, err := s.teamRepo.FindByUserID(ctx, userID)
	if errors.Is(err, pg.ErrTeamNotFound) {
		return nil, ErrTeamNotFound
	} else if err != nil {
		return nil, fmt.Errorf("find team for reviewer: %w", err)
	}

	var candidates []uuid.UUID
	for _, uid := range team.UserIDs {
		if uid == pr.AuthorID || contains(pr.ReviewersIDs, uid) {
			continue
		}
		u, err := s.userRepo.FindByID(ctx, uid)
		if errors.Is(err, pg.ErrUserNotFound) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("find user %s: %w", uid, err)
		}
		if u.IsActive {
			candidates = append(candidates, uid)
		}
	}

	if len(candidates) == 0 {
		return nil, ErrNoReviewCandidates
	}

	newReviewer := candidates[rand.Intn(len(candidates))]
	pr.ReviewersIDs[reviewerIndex] = newReviewer

	if err := s.prRepo.Update(ctx, pr); err != nil {
		return nil, fmt.Errorf("update PR: %w", err)
	}

	statusName := "OPEN"
	if status != nil {
		statusName = status.Name
	} else {
		if st, err := s.statusRepo.FindByID(ctx, pr.StatusID); err == nil && st != nil {
			statusName = st.Name
		}
	}

	return &dtos.ReassignReviewerResponse{
		Pr:         *dtos.ModelToPullRequestDTO(pr, statusName),
		ReplacedBy: newReviewer.String(),
	}, nil
}

func (s *PullRequestService) FindPullRequestsByReviewer(ctx context.Context, userID uuid.UUID) ([]*dtos.PullRequest, error) {
	prs, err := s.prRepo.FindByReviewer(ctx, userID)
	if err != nil {
		return nil, err
	}

	var dts []*dtos.PullRequest
	for _, pr := range prs {
		status, err := s.statusRepo.FindByID(ctx, pr.StatusID)
		if err != nil && !errors.Is(err, pg.ErrStatusNotFound) {
			return nil, fmt.Errorf("find pull request status %s: %w", pr.StatusID, err)
		}
		dto := dtos.ModelToPullRequestDTO(pr, status.Name)
		dts = append(dts, dto)
	}

	return dts, err
}

func (s *PullRequestService) MarkAsMerged(ctx context.Context, prID uuid.UUID) (*dtos.PullRequest, error) {
	pr, err := s.prRepo.FindByID(ctx, prID)
	if errors.Is(err, pg.ErrPullRequestNotFound) {
		return nil, ErrPRNotFound
	} else if err != nil {
		return nil, fmt.Errorf("find PR: %w", err)
	}

	mergedStatus, err := s.getMergedStatus(ctx)
	if err != nil {
		return nil, err
	}

	if pr.StatusID == mergedStatus.ID {
		return dtos.ModelToPullRequestDTO(pr, mergedStatus.Name), nil
	}

	now := time.Now()
	pr.StatusID = mergedStatus.ID
	pr.MergedAt = &now

	if err := s.prRepo.Update(ctx, pr); err != nil {
		return nil, fmt.Errorf("update PR to merged: %w", err)
	}

	return dtos.ModelToPullRequestDTO(pr, mergedStatus.Name), nil
}

func (s *PullRequestService) GetUserReviews(ctx context.Context, userID uuid.UUID) (*dtos.UserGetReviewResponse, error) {
	allPRs, err := s.prRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("find all PRs: %w", err)
	}

	var prs []*models.PullRequest
	for _, pr := range allPRs {
		if isReviewer(pr.ReviewersIDs, userID) {
			prs = append(prs, pr)
		}
	}

	pullRequests := make([]dtos.PullRequestShort, len(prs))

	for i, pr := range prs {
		status, err := s.statusRepo.FindByID(ctx, pr.StatusID)
		if err != nil {
			return nil, err
		}
		pullRequests[i] = dtos.PullRequestShort{
			PullRequestId:   pr.ID.String(),
			PullRequestName: pr.Title,
			AuthorId:        pr.AuthorID.String(),
			Status:          dtos.PullRequestStatus(status.Name),
		}
	}

	return &dtos.UserGetReviewResponse{
		UserId:       userID.String(),
		PullRequests: pullRequests,
	}, nil

}

func isReviewer(reviewers []uuid.UUID, userID uuid.UUID) bool {
	for _, r := range reviewers {
		if r == userID {
			return true
		}
	}
	return false
}

func (s *PullRequestService) CreatePullRequest(ctx context.Context, req *dtos.PullRequest) (*dtos.PullRequest, error) {
	prID := toUUID(req.PullRequestId)

	authorID, err := uuid.Parse(req.AuthorId)
	if err != nil {
		return nil, errors.New("invalid author_id format")
	}

	return s.CreateWithReviewers(ctx, prID, req.PullRequestName, authorID)
}

func toUUID(externalID string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(externalID))
}

func (s *PullRequestService) getOpenStatus(ctx context.Context) (*models.Status, error) {
	statuses, err := s.statusRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get statuses: %w", err)
	}
	for _, st := range statuses {
		if st.Name == "OPEN" {
			return st, nil
		}
	}
	return nil, fmt.Errorf("status 'OPEN' not found")
}

func (s *PullRequestService) getMergedStatus(ctx context.Context) (*models.Status, error) {
	statuses, err := s.statusRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get statuses: %w", err)
	}
	for _, st := range statuses {
		if st.Name == "MERGED" {
			return st, nil
		}
	}
	return nil, fmt.Errorf("status 'MERGED' not found")
}

func chooseRandomUsers(userIDs []uuid.UUID, max int) []uuid.UUID {
	n := len(userIDs)
	if n <= max {
		cpy := make([]uuid.UUID, len(userIDs))
		copy(cpy, userIDs)
		return cpy
	}
	selected := make([]uuid.UUID, 0, max)
	perm := rand.Perm(n)
	for i := 0; i < max; i++ {
		selected = append(selected, userIDs[perm[i]])
	}
	return selected
}

func contains(slice []uuid.UUID, item uuid.UUID) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
