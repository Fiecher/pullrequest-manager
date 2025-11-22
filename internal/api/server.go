package api

import (
	"errors"
	"fmt"
	"net/http"
	"pullrequest-inator/internal/api/dtos"
	"pullrequest-inator/internal/infrastructure/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Server struct {
	prService   *services.PullRequestService
	teamService *services.TeamService
	userService *services.UserService
}

func NewServer(prService *services.PullRequestService, teamService *services.TeamService, userService *services.UserService) (*Server, error) {
	if prService == nil {
		return nil, errors.New("prService is required")
	}
	if teamService == nil {
		return nil, errors.New("teamService is required")
	}
	if userService == nil {
		return nil, errors.New("userService is required")
	}

	return &Server{
		prService:   prService,
		teamService: teamService,
		userService: userService,
	}, nil
}

func (s *Server) PostPullRequestCreate(ctx echo.Context) error {
	var input PostPullRequestCreateJSONRequestBody
	if err := ctx.Bind(&input); err != nil {
		return err
	}

	dtoReq := &dtos.PullRequest{
		PullRequestId:   input.PullRequestId,
		PullRequestName: input.PullRequestName,
		AuthorId:        input.AuthorId,
	}

	pr, err := s.prService.CreatePullRequest(ctx.Request().Context(), dtoReq)
	if err != nil {
		return mapAppErrorToEchoResponse(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, map[string]any{
		"pr": ToAPIPullRequest(*pr),
	})
}

func (s *Server) PostPullRequestMerge(ctx echo.Context) error {
	var input PostPullRequestMergeJSONRequestBody
	if err := ctx.Bind(&input); err != nil {
		return err
	}

	prID, err := uuid.Parse(input.PullRequestId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid pull_request_id")
	}

	pr, err := s.prService.MarkAsMerged(ctx.Request().Context(), prID)
	if err != nil {
		return mapAppErrorToEchoResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"pr": ToAPIPullRequest(*pr),
	})
}

func (s *Server) PostPullRequestReassign(ctx echo.Context) error {
	var input PostPullRequestReassignJSONRequestBody
	if err := ctx.Bind(&input); err != nil {
		return err
	}

	prID, err := uuid.Parse(input.PullRequestId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid pull_request_id")
	}

	oldID, err := uuid.Parse(input.OldUserId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid old_user_id")
	}

	resp, err := s.prService.ReassignReviewer(ctx.Request().Context(), oldID, prID)
	if err != nil {
		return mapAppErrorToEchoResponse(ctx, err)
	}
	updatedPR := resp.Pr
	replacedBy := resp.ReplacedBy

	return ctx.JSON(http.StatusOK, map[string]any{
		"pr":          ToAPIPullRequest(updatedPR),
		"replaced_by": replacedBy,
	})
}

func (s *Server) PostTeamAdd(ctx echo.Context) error {
	var team Team
	if err := ctx.Bind(&team); err != nil {
		return err
	}

	dtoTeam := FromAPITeam(team)
	err := s.teamService.CreateTeamWithUsers(ctx.Request().Context(), &dtoTeam)
	if err != nil {
		return mapAppErrorToEchoResponse(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, map[string]any{
		"team": team,
	})
}

func (s *Server) GetTeamGet(ctx echo.Context, params GetTeamGetParams) error {
	team, err := s.teamService.GetTeamByName(ctx.Request().Context(), params.TeamName)
	if err != nil {
		return mapAppErrorToEchoResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, ToAPITeam(*team))
}

func (s *Server) PostUsersSetIsActive(ctx echo.Context) error {
	var input PostUsersSetIsActiveJSONRequestBody
	if err := ctx.Bind(&input); err != nil {
		return err
	}

	userID, err := uuid.Parse(input.UserId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}

	updated, err := s.teamService.SetUserActiveByID(ctx.Request().Context(), userID, input.IsActive)
	if err != nil {
		return mapAppErrorToEchoResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, map[string]User{
		"user": ToAPIUser(*updated),
	})
}

func (s *Server) GetUsersGetReview(ctx echo.Context, params GetUsersGetReviewParams) error {
	userID, err := uuid.Parse(params.UserId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}

	resp, err := s.prService.FindPullRequestsByReviewer(ctx.Request().Context(), userID)
	if err != nil {
		return mapAppErrorToEchoResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"user_id":       params.UserId,
		"pull_requests": ToAPIPullRequestShortList(resp),
	})
}

func mapAppErrorToEchoResponse(ctx echo.Context, err error) error {
	switch {
	case errors.Is(err, services.ErrPRAlreadyExists):
		return ctx.JSON(http.StatusConflict, map[string]string{
			"error": "pull request already exists",
		})

	case errors.Is(err, services.ErrAuthorNotFound):
		return ctx.JSON(http.StatusNotFound, map[string]string{
			"error": "author not found",
		})

	case errors.Is(err, services.ErrTeamNotFound):
		return ctx.JSON(http.StatusNotFound, map[string]string{
			"error": "team not found",
		})

	case errors.Is(err, services.ErrUserNotReviewer):
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "user is not a reviewer",
		})

	case errors.Is(err, services.ErrPRAlreadyMerged):
		return ctx.JSON(http.StatusConflict, map[string]string{
			"error": "pull request already merged",
		})

	case errors.Is(err, services.ErrNoReviewCandidates):
		return ctx.JSON(http.StatusConflict, map[string]string{
			"error": "no active users to assign as reviewers",
		})

	case errors.Is(err, services.ErrTeamExists):
		return ctx.JSON(http.StatusConflict, map[string]string{
			"error": "team already exists",
		})
	}

	return ctx.JSON(http.StatusInternalServerError, map[string]any{
		"error":          "internal server error",
		"original_error": err.Error(),
		"error_type":     fmt.Sprintf("%T", err),
	})
}
