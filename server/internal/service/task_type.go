package service

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"
	"dcs/pkg/errors"
	"time"

	"github.com/google/uuid"
)

type TaskTypeService struct {
	repo repository.TaskTypeRepo
}

func NewTaskTypeService(repo repository.TaskTypeRepo) *TaskTypeService {
	return &TaskTypeService{repo: repo}
}

func (s *TaskTypeService) Create(ctx context.Context, tt *model.TaskType) (*model.TaskType, error) {
	if tt.Name == "" {
		return nil, errors.Wrap(nil, 40001, 400, "name is required")
	}
	if tt.SourceType == "" {
		return nil, errors.Wrap(nil, 40001, 400, "source_type is required")
	}
	if tt.Collector == "" {
		return nil, errors.Wrap(nil, 40001, 400, "collector is required")
	}

	tt.TypeID = "tt-" + uuid.New().String()[:8]
	tt.Version = 1
	tt.CreatedAt = time.Now()
	tt.UpdatedAt = time.Now()

	if err := s.repo.Create(ctx, tt); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to create task type")
	}

	return tt, nil
}

func (s *TaskTypeService) GetByID(ctx context.Context, id string) (*model.TaskType, error) {
	tt, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get task type")
	}
	if tt == nil {
		return nil, errors.ErrNotFound
	}
	return tt, nil
}

func (s *TaskTypeService) List(ctx context.Context) ([]*model.TaskType, error) {
	return s.repo.List(ctx)
}

func (s *TaskTypeService) Update(ctx context.Context, id string, updates *model.TaskType) (*model.TaskType, error) {
	tt, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get task type")
	}
	if tt == nil {
		return nil, errors.ErrNotFound
	}

	if updates.Name != "" {
		tt.Name = updates.Name
	}
	if updates.Description != "" {
		tt.Description = updates.Description
	}
	if updates.SourceType != "" {
		tt.SourceType = updates.SourceType
	}
	if updates.Collector != "" {
		tt.Collector = updates.Collector
	}
	if updates.FormSchema != "" {
		tt.FormSchema = updates.FormSchema
	}
	if updates.DetailPanels != "" {
		tt.DetailPanels = updates.DetailPanels
	}
	if updates.DeclaredMetrics != "" {
		tt.DeclaredMetrics = updates.DeclaredMetrics
	}
	if updates.ConfigTemplate != "" {
		tt.ConfigTemplate = updates.ConfigTemplate
	}
	tt.Version = updates.Version
	tt.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, tt); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to update task type")
	}

	return tt, nil
}

func (s *TaskTypeService) Delete(ctx context.Context, id string) error {
	tt, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, 50001, 500, "failed to get task type")
	}
	if tt == nil {
		return errors.ErrNotFound
	}

	if tt.IsBuiltin {
		return errors.Wrap(nil, 42202, 422, "builtin task types cannot be deleted")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return errors.Wrap(err, 50001, 500, "failed to delete task type")
	}

	return nil
}
