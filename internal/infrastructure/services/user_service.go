package services

import (
	"context"
	"errors"
	"pullrequest-inator/internal/api/dtos"
	"pullrequest-inator/internal/infrastructure/models"
	"pullrequest-inator/internal/infrastructure/repositories/interfaces"

	"github.com/google/uuid"
)

type UserService struct {
	userRepo repositories.User
}

func NewUserService(userRepo repositories.User) (*UserService, error) {
	if userRepo == nil {
		return nil, errors.New("userRepository cannot be nil")
	}
	return &UserService{userRepo: userRepo}, nil
}

func (s *UserService) RegisterUser(ctx context.Context, user *dtos.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}

	userID, err := uuid.Parse(user.UserId)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	userModel := &models.User{
		ID:       userID,
		Username: user.Username,
		IsActive: user.IsActive,
	}

	return s.userRepo.Create(ctx, userModel)
}

func (s *UserService) UnregisterUserByID(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.DeleteByID(ctx, userID)
}

func (s *UserService) ListUsers(ctx context.Context) ([]*dtos.User, error) {
	users, err := s.userRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	apiUsers := make([]*dtos.User, len(users))
	for i, u := range users {
		apiUsers[i] = &dtos.User{
			UserId:   u.ID.String(),
			Username: u.Username,
			IsActive: u.IsActive,
		}
	}
	return apiUsers, nil
}

func (s *UserService) SetUserActive(ctx context.Context, userID uuid.UUID, active bool) (*dtos.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.IsActive = active
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return &dtos.User{
		UserId:   user.ID.String(),
		Username: user.Username,
		IsActive: user.IsActive,
	}, nil
}
