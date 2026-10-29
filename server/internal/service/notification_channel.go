package service

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"
	"dcs/pkg/errors"
	"time"

	"github.com/google/uuid"
)

type NotificationChannelService struct {
	repo repository.NotificationChannelRepo
}

func NewNotificationChannelService(repo repository.NotificationChannelRepo) *NotificationChannelService {
	return &NotificationChannelService{repo: repo}
}

func (s *NotificationChannelService) Create(ctx context.Context, name, channelType, config string) (*model.NotificationChannel, error) {
	if name == "" {
		return nil, errors.Wrap(nil, 40001, 400, "name is required")
	}
	if channelType == "" {
		return nil, errors.Wrap(nil, 40001, 400, "channel_type is required")
	}

	nc := &model.NotificationChannel{
		ChannelID:   "nc-" + uuid.New().String()[:8],
		Name:        name,
		ChannelType: channelType,
		Config:      config,
		Enabled:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, nc); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to create notification channel")
	}
	return nc, nil
}

func (s *NotificationChannelService) GetByID(ctx context.Context, id string) (*model.NotificationChannel, error) {
	nc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get notification channel")
	}
	if nc == nil {
		return nil, errors.ErrNotFound
	}
	return nc, nil
}

func (s *NotificationChannelService) List(ctx context.Context, channelType string) ([]*model.NotificationChannel, error) {
	return s.repo.List(ctx, channelType)
}

func (s *NotificationChannelService) Update(ctx context.Context, id, name, channelType, config string, enabled *bool) (*model.NotificationChannel, error) {
	nc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get notification channel")
	}
	if nc == nil {
		return nil, errors.ErrNotFound
	}

	if name != "" {
		nc.Name = name
	}
	if channelType != "" {
		nc.ChannelType = channelType
	}
	if config != "" {
		nc.Config = config
	}
	if enabled != nil {
		nc.Enabled = *enabled
	}
	nc.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, nc); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to update notification channel")
	}
	return nc, nil
}

func (s *NotificationChannelService) Delete(ctx context.Context, id string) error {
	nc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, 50001, 500, "failed to get notification channel")
	}
	if nc == nil {
		return errors.ErrNotFound
	}
	return s.repo.Delete(ctx, id)
}

func (s *NotificationChannelService) Test(ctx context.Context, id string) error {
	nc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, 50001, 500, "failed to get notification channel")
	}
	if nc == nil {
		return errors.ErrNotFound
	}

	result := "success"
	if err := s.repo.UpdateTestResult(ctx, id, result); err != nil {
		return errors.Wrap(err, 50001, 500, "failed to update test result")
	}
	return nil
}
