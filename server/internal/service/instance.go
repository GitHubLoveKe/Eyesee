package service

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"
	"dcs/pkg/errors"
	"time"

	"github.com/google/uuid"
)

type InstanceService struct {
	repo repository.InstanceRepo
}

func NewInstanceService(repo repository.InstanceRepo) *InstanceService {
	return &InstanceService{repo: repo}
}

func (s *InstanceService) Create(ctx context.Context, inst *model.Instance) (*model.Instance, error) {
	if inst.Name == "" {
		return nil, errors.Wrap(nil, 40001, 400, "name is required")
	}
	if inst.ZoneID == "" {
		return nil, errors.Wrap(nil, 40001, 400, "zone_id is required")
	}
	if inst.ResourceTypeID == "" {
		return nil, errors.Wrap(nil, 40001, 400, "resource_type_id is required")
	}
	if inst.Host == "" {
		return nil, errors.Wrap(nil, 40001, 400, "host is required")
	}

	inst.InstanceID = "inst-" + uuid.New().String()[:8]
	inst.Status = "pending"
	inst.CreatedAt = time.Now()
	inst.UpdatedAt = time.Now()

	if err := s.repo.Create(ctx, inst); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to create instance")
	}

	return inst, nil
}

func (s *InstanceService) GetByID(ctx context.Context, id string) (*model.Instance, error) {
	inst, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get instance")
	}
	if inst == nil {
		return nil, errors.ErrNotFound
	}
	return inst, nil
}

func (s *InstanceService) List(ctx context.Context, filter repository.InstanceFilter) ([]*model.Instance, int64, error) {
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return s.repo.List(ctx, filter)
}

func (s *InstanceService) Update(ctx context.Context, id string, updates *model.Instance) (*model.Instance, error) {
	inst, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get instance")
	}
	if inst == nil {
		return nil, errors.ErrNotFound
	}

	if updates.Name != "" {
		inst.Name = updates.Name
	}
	if updates.Host != "" {
		inst.Host = updates.Host
	}
	if updates.Port != nil {
		inst.Port = updates.Port
	}
	if updates.Labels != "" {
		inst.Labels = updates.Labels
	}
	if updates.Extra != "" {
		inst.Extra = updates.Extra
	}
	if updates.Status != "" {
		inst.Status = updates.Status
	}
	inst.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, inst); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to update instance")
	}

	return inst, nil
}

func (s *InstanceService) Delete(ctx context.Context, id string) error {
	inst, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, 50001, 500, "failed to get instance")
	}
	if inst == nil {
		return errors.ErrNotFound
	}

	has, err := s.repo.HasTasks(ctx, id)
	if err != nil {
		return errors.Wrap(err, 50001, 500, "failed to check tasks")
	}
	if has {
		return errors.Wrap(nil, 42202, 422, "has collection tasks, cannot delete")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return errors.Wrap(err, 50001, 500, "failed to delete instance")
	}

	return nil
}
