package services

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"pullrequest-inator/internal/domain/models"
	"pullrequest-inator/internal/infrastructure/database/pg"
	"pullrequest-inator/internal/infrastructure/dtos"
	"pullrequest-inator/internal/infrastructure/repositories"
	"time"

	"github.com/google/uuid"
)

var (
	ErrAuthorNotFound     = errors.New("author not found")
	ErrTeamNotFound       = errors.New("team not found")
	ErrPRAlreadyExists    = errors.New("pull request already exists")
	ErrPRNotFound         = errors.New("pull request not found")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserNotReviewer    = errors.New("user is not a reviewer")
	ErrNoReviewCandidates = errors.New("no users available to review")
	ErrPRAlreadyMerged    = errors.New("cannot change PR state because already merged")
)

type DefaultPullRequestService struct {
	userRepo   repositories.User
	prRepo     repositories.PullRequest
	teamRepo   repositories.Team
	statusRepo repositories.Status
}

func NewDefaultPullRequestService(
	userRepo repositories.User,
	prRepo repositories.PullRequest,
	teamRepo repositories.Team,
	statusRepo repositories.Status,
) (*DefaultPullRequestService, error) {
	return &DefaultPullRequestService{
		userRepo:   userRepo,
		prRepo:     prRepo,
		teamRepo:   teamRepo,
		statusRepo: statusRepo,
	}, nil
}

func (s *DefaultPullRequestService) CreateWithReviewers(ctx context.Context, prID uuid.UUID, prName string, authorID uuid.UUID) (*dtos.PullRequestDTO, error) {
	existing, err := s.prRepo.FindByID(ctx, prID)
	if err != nil && !errors.Is(err, pg.ErrPullRequestNotFound) {
		return nil, fmt.Errorf("check for existing PR: %w", err)
	}
	if existing != nil {
		return nil, ErrPRAlreadyExists
	}

	_, err = s.userRepo.FindByID(ctx, authorID)
	if errors.Is(err, pg.ErrUserNotFound) {
		return nil, ErrAuthorNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find author: %w", err)
	}

	team, err := s.teamRepo.FindByUserID(ctx, authorID)
	if errors.Is(err, pg.ErrTeamNotFound) {
		return nil, ErrTeamNotFound
	}
	if err != nil {
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
		}
		if err != nil {
			return nil, fmt.Errorf("find user %s for team %s: %w", uid, team.ID, err)
		}
		if u.IsActive {
			activeUsers = append(activeUsers, uid)
		}
	}

	if len(activeUsers) == 0 {
		return nil, ErrNoReviewCandidates
	}

	reviewers := chooseRandomUsers(activeUsers, 2)

	statuses, err := s.statusRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all statuses: %w", err)
	}
	var openStatusID uuid.UUID
	for _, st := range statuses {
		if st.Name == "OPEN" {
			openStatusID = st.ID
			break
		}
	}
	if openStatusID == uuid.Nil {
		return nil, fmt.Errorf("status 'OPEN' not found in database")
	}

	newPR := &models.PullRequest{
		ID:           prID,
		Title:        prName,
		AuthorID:     authorID,
		StatusID:     openStatusID,
		MergedAt:     nil,
		ReviewersIDs: reviewers,
	}

	if err := s.prRepo.Create(ctx, newPR); err != nil {
		return nil, fmt.Errorf("create pull request: %w", err)
	}

	return convertPullRequestToDTO(newPR, statuses), nil
}

func (s *DefaultPullRequestService) ReassignReviewer(ctx context.Context, userID uuid.UUID, prID uuid.UUID) (*dtos.ReassignReviewerResponseDTO, error) {
	pr, err := s.prRepo.FindByID(ctx, prID)
	if errors.Is(err, pg.ErrPullRequestNotFound) {
		return nil, ErrPRNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find PR for reassignment: %w", err)
	}

	status, err := s.statusRepo.FindByID(ctx, pr.StatusID)
	if err != nil && !errors.Is(err, pg.ErrStatusNotFound) {
		return nil, fmt.Errorf("find status for PR: %w", err)
	}
	if status != nil && status.Name == "MERGED" {
		return nil, ErrPRAlreadyMerged
	}

	isReviewer := false
	reviewerIndex := -1
	for i, rid := range pr.ReviewersIDs {
		if rid == userID {
			isReviewer = true
			reviewerIndex = i
			break
		}
	}
	if !isReviewer {
		return nil, ErrUserNotReviewer
	}

	team, err := s.teamRepo.FindByUserID(ctx, userID)
	if errors.Is(err, pg.ErrTeamNotFound) {
		return nil, ErrTeamNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find team for reviewer %s: %w", userID, err)
	}

	var candidates []uuid.UUID
	for _, uid := range team.UserIDs {
		if uid == pr.AuthorID || contains(pr.ReviewersIDs, uid) {
			continue
		}
		u, err := s.userRepo.FindByID(ctx, uid)
		if errors.Is(err, pg.ErrUserNotFound) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("find user %s for team %s: %w", uid, team.ID, err)
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
		return nil, fmt.Errorf("update pull request after reassignment: %w", err)
	}

	statuses, err := s.statusRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get statuses for response DTO: %w", err)
	}
	prDTO := convertPullRequestToDTO(pr, statuses)

	return &dtos.ReassignReviewerResponseDTO{
		Pr:         *prDTO,
		ReplacedBy: newReviewer,
	}, nil
}

func (s *DefaultPullRequestService) MarkAsMerged(ctx context.Context, prID uuid.UUID) (*dtos.PullRequestDTO, error) {
	pr, err := s.prRepo.FindByID(ctx, prID)
	if errors.Is(err, pg.ErrPullRequestNotFound) {
		return nil, ErrPRNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find PR to mark as merged: %w", err)
	}

	statuses, err := s.statusRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all statuses: %w", err)
	}
	var mergedStatusID uuid.UUID
	for _, st := range statuses {
		if st.Name == "MERGED" {
			mergedStatusID = st.ID
			break
		}
	}
	if mergedStatusID == uuid.Nil {
		return nil, fmt.Errorf("status 'MERGED' not found in database")
	}

	if pr.StatusID == mergedStatusID {
		return convertPullRequestToDTO(pr, statuses), nil
	}

	now := time.Now()
	pr.StatusID = mergedStatusID
	pr.MergedAt = &now

	if err := s.prRepo.Update(ctx, pr); err != nil {
		return nil, fmt.Errorf("update pull request to merged: %w", err)
	}

	return convertPullRequestToDTO(pr, statuses), nil
}

func (s *DefaultPullRequestService) CreateTeam(ctx context.Context, teamName string, members []dtos.TeamMemberDTO) error {
	existingTeam, err := s.teamRepo.FindByName(ctx, teamName)
	if err != nil && !errors.Is(err, pg.ErrTeamNotFound) {
		return fmt.Errorf("check for existing team: %w", err)
	}
	if existingTeam != nil {
		team := &models.Team{
			ID:      existingTeam.ID,
			Name:    teamName,
			UserIDs: []uuid.UUID{},
		}
		for _, member := range members {
			user, err := s.userRepo.FindByID(ctx, member.UserID)
			if errors.Is(err, pg.ErrUserNotFound) {
				newUser := &models.User{
					ID:       member.UserID,
					Username: member.Username,
					IsActive: member.IsActive,
				}
				if err := s.userRepo.Create(ctx, newUser); err != nil {
					return fmt.Errorf("create user %s: %w", member.UserID, err)
				}
				team.UserIDs = append(team.UserIDs, newUser.ID)
			} else if err != nil {
				return fmt.Errorf("find user %s: %w", member.UserID, err)
			} else {
				user.Username = member.Username
				user.IsActive = member.IsActive
				if err := s.userRepo.Update(ctx, user); err != nil {
					return fmt.Errorf("update user %s: %w", member.UserID, err)
				}
				team.UserIDs = append(team.UserIDs, user.ID)
			}
		}
		if err := s.teamRepo.Update(ctx, team); err != nil {
			return fmt.Errorf("update team: %w", err)
		}
		return nil
	}

	team := &models.Team{
		Name:    teamName,
		UserIDs: []uuid.UUID{},
	}

	for _, member := range members {
		user, err := s.userRepo.FindByID(ctx, member.UserID)
		if errors.Is(err, pg.ErrUserNotFound) {
			newUser := &models.User{
				ID:       member.UserID,
				Username: member.Username,
				IsActive: member.IsActive,
			}
			if err := s.userRepo.Create(ctx, newUser); err != nil {
				return fmt.Errorf("create user %s: %w", member.UserID, err)
			}
			team.UserIDs = append(team.UserIDs, newUser.ID)
		} else if err != nil {
			return fmt.Errorf("find user %s: %w", member.UserID, err)
		} else {
			user.Username = member.Username
			user.IsActive = member.IsActive
			if err := s.userRepo.Update(ctx, user); err != nil {
				return fmt.Errorf("update user %s: %w", member.UserID, err)
			}
			team.UserIDs = append(team.UserIDs, user.ID)
		}
	}

	if err := s.teamRepo.Create(ctx, team); err != nil {
		return fmt.Errorf("create team: %w", err)
	}

	return nil
}

func (s *DefaultPullRequestService) GetTeam(ctx context.Context, teamName string) (*dtos.TeamDTO, error) {
	team, err := s.teamRepo.FindByName(ctx, teamName)
	if errors.Is(err, pg.ErrTeamNotFound) {
		return nil, ErrTeamNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find team by name: %w", err)
	}

	membersDTO := make([]dtos.TeamMemberDTO, len(team.UserIDs))
	for i, userID := range team.UserIDs {
		user, err := s.userRepo.FindByID(ctx, userID)
		if errors.Is(err, pg.ErrUserNotFound) {
			membersDTO[i] = dtos.TeamMemberDTO{
				UserID:   userID,
				Username: "",
				IsActive: false,
			}
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("find user %s for team %s: %w", userID, team.ID, err)
		}
		membersDTO[i] = dtos.TeamMemberDTO{
			UserID:   user.ID,
			Username: user.Username,
			IsActive: user.IsActive,
		}
	}

	return &dtos.TeamDTO{
		TeamName: teamName,
		Members:  membersDTO,
	}, nil
}

func (s *DefaultPullRequestService) SetUserActive(ctx context.Context, userID uuid.UUID, isActive bool) (*dtos.UserDTO, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if errors.Is(err, pg.ErrUserNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user to update: %w", err)
	}

	user.IsActive = isActive
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	team, err := s.teamRepo.FindByUserID(ctx, userID)
	var teamName string
	if errors.Is(err, pg.ErrTeamNotFound) {
		teamName = ""
	} else if err != nil {
		return nil, fmt.Errorf("find team for user: %w", err)
	} else {
		teamName = team.Name
	}

	return &dtos.UserDTO{
		UserID:   user.ID,
		Username: user.Username,
		TeamName: teamName,
		IsActive: user.IsActive,
	}, nil
}

func (s *DefaultPullRequestService) GetUserReviews(ctx context.Context, userID uuid.UUID) (*dtos.UserGetReviewResponseDTO, error) {
	allPRs, err := s.prRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("find all PRs to filter by reviewer: %w", err)
	}

	var relevantPRs []dtos.PullRequestShortDTO
	statusMap := make(map[uuid.UUID]string)

	for _, pr := range allPRs {
		isReviewer := false
		for _, rid := range pr.ReviewersIDs {
			if rid == userID {
				isReviewer = true
				break
			}
		}
		if !isReviewer {
			continue
		}

		statusName, ok := statusMap[pr.StatusID]
		if !ok {
			status, err := s.statusRepo.FindByID(ctx, pr.StatusID)
			if err != nil && !errors.Is(err, pg.ErrStatusNotFound) {
				return nil, fmt.Errorf("find status for PR %s: %w", pr.ID, err)
			}
			if status != nil {
				statusName = status.Name
			}
			statusMap[pr.StatusID] = statusName
		}

		relevantPRs = append(relevantPRs, dtos.PullRequestShortDTO{
			PullRequestID:   pr.ID,
			PullRequestName: pr.Title,
			AuthorID:        pr.AuthorID,
			Status:          statusName,
		})
	}

	return &dtos.UserGetReviewResponseDTO{
		UserID:       userID,
		PullRequests: relevantPRs,
	}, nil
}

func chooseRandomUsers(userIDs []uuid.UUID, max int) []uuid.UUID {
	n := len(userIDs)
	if n <= max {
		return append([]uuid.UUID{}, userIDs...)
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

func convertStatusToStringMap(statuses []*models.Status) map[uuid.UUID]string {
	m := make(map[uuid.UUID]string, len(statuses))
	for _, st := range statuses {
		m[st.ID] = st.Name
	}
	return m
}

func convertPullRequestToDTO(pr *models.PullRequest, statuses []*models.Status) *dtos.PullRequestDTO {
	statusMap := convertStatusToStringMap(statuses)
	statusName := statusMap[pr.StatusID]

	return &dtos.PullRequestDTO{
		PullRequestID:     pr.ID,
		PullRequestName:   pr.Title,
		AuthorID:          pr.AuthorID,
		Status:            statusName,
		AssignedReviewers: pr.ReviewersIDs,
		CreatedAt:         &pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}
