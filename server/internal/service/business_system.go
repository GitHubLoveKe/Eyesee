package service

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"
	"dcs/pkg/errors"
	"time"

	"github.com/google/uuid"
)

type BusinessSystemService struct {
	repo repository.BusinessSystemRepo
}

func NewBusinessSystemService(repo repository.BusinessSystemRepo) *BusinessSystemService {
	return &BusinessSystemService{repo: repo}
}

func (s *BusinessSystemService) Create(ctx context.Context, name, description string) (*model.BusinessSystem, error) {
	if name == "" {
		return nil, errors.Wrap(nil, 40001, 400, "name is required")
	}

	bs := &model.BusinessSystem{
		SystemID:    "bs-" + uuid.New().String()[:8],
		Name:        name,
		Description: description,
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, bs); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to create business system")
	}

	return bs, nil
}

func (s *BusinessSystemService) GetByID(ctx context.Context, id string) (*model.BusinessSystem, error) {
	bs, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get business system")
	}
	if bs == nil {
		return nil, errors.ErrNotFound
	}
	return bs, nil
}

func (s *BusinessSystemService) List(ctx context.Context, keyword string, page, pageSize int) ([]*model.BusinessSystem, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, keyword, offset, pageSize)
}

func (s *BusinessSystemService) Update(ctx context.Context, id, name, description, status string) (*model.BusinessSystem, error) {
	bs, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get business system")
	}
	if bs == nil {
		return nil, errors.ErrNotFound
	}

	if name != "" {
		bs.Name = name
	}
	if description != "" {
		bs.Description = description
	}
	if status != "" {
		bs.Status = status
	}
	bs.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, bs); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to update business system")
	}

	return bs, nil
}

func (s *BusinessSystemService) Delete(ctx context.Context, id string) error {
	bs, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, 50001, 500, "failed to get business system")
	}
	if bs == nil {
		return errors.ErrNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return errors.Wrap(err, 50001, 500, "failed to delete business system")
	}

	return nil
}

func (s *BusinessSystemService) AddMember(ctx context.Context, systemID, userID, role string) error {
	if role == "" {
		role = "viewer"
	}

	m := &model.BusinessSystemMember{
		SystemID:  systemID,
		UserID:    userID,
		Role:      role,
		CreatedAt: time.Now(),
	}

	return s.repo.AddMember(ctx, m)
}

func (s *BusinessSystemService) RemoveMember(ctx context.Context, systemID, userID string) error {
	return s.repo.RemoveMember(ctx, systemID, userID)
}

func (s *BusinessSystemService) ListMembers(ctx context.Context, systemID string) ([]*model.BusinessSystemMember, error) {
	return s.repo.ListMembers(ctx, systemID)
}
