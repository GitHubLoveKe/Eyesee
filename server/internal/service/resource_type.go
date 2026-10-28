package service

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"
	"dcs/pkg/errors"
	"time"

	"github.com/google/uuid"
)

type ResourceSystemService struct {
	repo repository.ResourceTypeRepo
}

func NewResourceSystemService(repo repository.ResourceTypeRepo) *ResourceSystemService {
	return &ResourceSystemService{repo: repo}
}

func (s *ResourceSystemService) Create(ctx context.Context, parentID *string, name, description, icon string, sortOrder int) (*model.ResourceType, error) {
	if name == "" {
		return nil, errors.Wrap(nil, 40001, 400, "name is required")
	}

	rt := &model.ResourceType{
		TypeID:      "rt-" + uuid.New().String()[:8],
		ParentID:    parentID,
		Name:        name,
		Description: description,
		Icon:        icon,
		SortOrder:   sortOrder,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, rt); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to create resource type")
	}

	return rt, nil
}

func (s *ResourceSystemService) GetByID(ctx context.Context, id string) (*model.ResourceType, error) {
	rt, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get resource type")
	}
	if rt == nil {
		return nil, errors.ErrNotFound
	}
	return rt, nil
}

func (s *ResourceSystemService) List(ctx context.Context, parentID *string) ([]*model.ResourceType, error) {
	return s.repo.List(ctx, parentID)
}

func (s *ResourceSystemService) GetTree(ctx context.Context) ([]*model.ResourceType, error) {
	return s.repo.List(ctx, nil)
}

func (s *ResourceSystemService) Update(ctx context.Context, id string, parentID *string, name, description, icon string, sortOrder int) (*model.ResourceType, error) {
	rt, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get resource type")
	}
	if rt == nil {
		return nil, errors.ErrNotFound
	}

	if name != "" {
		rt.Name = name
	}
	rt.ParentID = parentID
	rt.Description = description
	rt.Icon = icon
	rt.SortOrder = sortOrder
	rt.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, rt); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to update resource type")
	}

	return rt, nil
}

func (s *ResourceSystemService) Delete(ctx context.Context, id string) error {
	has, err := s.repo.HasChildren(ctx, id)
	if err != nil {
		return errors.Wrap(err, 50001, 500, "failed to check children")
	}
	if has {
		return errors.Wrap(nil, 42202, 422, "has child types, cannot delete")
	}

	has, err = s.repo.HasInstances(ctx, id)
	if err != nil {
		return errors.Wrap(err, 50001, 500, "failed to check instances")
	}
	if has {
		return errors.Wrap(nil, 42202, 422, "has linked instances, cannot delete")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return errors.Wrap(err, 50001, 500, "failed to delete resource type")
	}

	return nil
}
