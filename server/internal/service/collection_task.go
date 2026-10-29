package service

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"
	"dcs/pkg/errors"
	"time"

	"github.com/google/uuid"
)

type CollectionTaskService struct {
	repo repository.CollectionTaskRepo
}

func NewCollectionTaskService(repo repository.CollectionTaskRepo) *CollectionTaskService {
	return &CollectionTaskService{repo: repo}
}

func (s *CollectionTaskService) Create(ctx context.Context, ct *model.CollectionTask) (*model.CollectionTask, error) {
	if ct.Name == "" {
		return nil, errors.Wrap(nil, 40001, 400, "name is required")
	}
	if ct.InstanceID == "" {
		return nil, errors.Wrap(nil, 40001, 400, "instance_id is required")
	}
	if ct.TypeID == "" {
		return nil, errors.Wrap(nil, 40001, 400, "type_id is required")
	}

	ct.TaskID = "ct-" + uuid.New().String()[:8]
	ct.Status = "pending"
	ct.Enabled = true
	ct.CreatedAt = time.Now()
	ct.UpdatedAt = time.Now()

	if ct.Module == "" {
		ct.Module = "metrics"
	}

	if err := s.repo.Create(ctx, ct); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to create collection task")
	}

	return ct, nil
}

func (s *CollectionTaskService) GetByID(ctx context.Context, id string) (*model.CollectionTask, error) {
	ct, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get collection task")
	}
	if ct == nil {
		return nil, errors.ErrNotFound
	}
	return ct, nil
}

func (s *CollectionTaskService) List(ctx context.Context, filter repository.CollectionTaskFilter) ([]*model.CollectionTask, int64, error) {
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return s.repo.List(ctx, filter)
}

func (s *CollectionTaskService) Update(ctx context.Context, id string, updates *model.CollectionTask) (*model.CollectionTask, error) {
	ct, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get collection task")
	}
	if ct == nil {
		return nil, errors.ErrNotFound
	}

	if updates.Name != "" {
		ct.Name = updates.Name
	}
	if updates.Config != "" {
		ct.Config = updates.Config
	}
	if updates.Status != "" {
		ct.Status = updates.Status
	}
	ct.Enabled = updates.Enabled
	ct.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, ct); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to update collection task")
	}

	return ct, nil
}

func (s *CollectionTaskService) Delete(ctx context.Context, id string) error {
	ct, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, 50001, 500, "failed to get collection task")
	}
	if ct == nil {
		return errors.ErrNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return errors.Wrap(err, 50001, 500, "failed to delete collection task")
	}

	return nil
}

func (s *CollectionTaskService) Enable(ctx context.Context, id string) error {
	ct, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, 50001, 500, "failed to get collection task")
	}
	if ct == nil {
		return errors.ErrNotFound
	}

	ct.Enabled = true
	ct.UpdatedAt = time.Now()
	return s.repo.Update(ctx, ct)
}

func (s *CollectionTaskService) Disable(ctx context.Context, id string) error {
	ct, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, 50001, 500, "failed to get collection task")
	}
	if ct == nil {
		return errors.ErrNotFound
	}

	ct.Enabled = false
	ct.UpdatedAt = time.Now()
	return s.repo.Update(ctx, ct)
}
