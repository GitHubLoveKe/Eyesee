package service

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"
	"dcs/pkg/errors"
	"time"
)

type UserAuthService struct {
	userRepo   repository.UserRepo
	targetRepo repository.UserNotificationTargetRepo
}

func NewUserAuthService(userRepo repository.UserRepo, targetRepo repository.UserNotificationTargetRepo) *UserAuthService {
	return &UserAuthService{userRepo: userRepo, targetRepo: targetRepo}
}

func (s *UserAuthService) ChangePassword(ctx context.Context, userID, newPassword string) error {
	// In a real implementation, this would hash the password and update user_auth_source
	// For now, this is a placeholder that validates the user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.Wrap(err, 50001, 500, "failed to get user")
	}
	if user == nil {
		return errors.ErrNotFound
	}
	// Password update would go through auth_source table
	return nil
}

func (s *UserAuthService) ListNotificationTargets(ctx context.Context, userID string) ([]*model.UserNotificationTarget, error) {
	return s.targetRepo.ListByUserID(ctx, userID)
}

func (s *UserAuthService) AddNotificationTarget(ctx context.Context, req AddNotificationTargetRequest) (*model.UserNotificationTarget, error) {
	if req.TargetValue == "" {
		return nil, errors.Wrap(nil, 40001, 400, "target_value is required")
	}

	t := &model.UserNotificationTarget{
		UserID:      req.UserID,
		ChannelType: req.ChannelType,
		TargetValue: req.TargetValue,
		IsPrimary:   req.IsPrimary,
		CreatedAt:   time.Now(),
	}
	if err := s.targetRepo.Create(ctx, t); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to create notification target")
	}
	return t, nil
}

type AddNotificationTargetRequest struct {
	UserID      string
	ChannelType string
	TargetValue string
	IsPrimary   bool
}

func (s *UserAuthService) RemoveNotificationTarget(ctx context.Context, id int64) error {
	return s.targetRepo.Delete(ctx, id)
}

func (s *UserAuthService) SetPrimaryTarget(ctx context.Context, userID string, id int64) error {
	return s.targetRepo.SetPrimary(ctx, userID, id)
}
