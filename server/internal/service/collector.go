package service

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"
	"dcs/pkg/errors"
	"time"

	"github.com/google/uuid"
)

type CollectorService struct {
	repo repository.CollectorNodeRepo
}

func NewCollectorService(repo repository.CollectorNodeRepo) *CollectorService {
	return &CollectorService{repo: repo}
}

func (s *CollectorService) Register(ctx context.Context, n *model.CollectorNode) (*model.CollectorNode, error) {
	if n.Name == "" {
		return nil, errors.Wrap(nil, 40001, 400, "name is required")
	}
	if n.ZoneID == "" {
		return nil, errors.Wrap(nil, 40001, 400, "zone_id is required")
	}
	if n.Address == "" {
		return nil, errors.Wrap(nil, 40001, 400, "address is required")
	}

	n.NodeID = "node-" + uuid.New().String()[:8]
	n.Status = "online"
	n.CreatedAt = time.Now()
	n.UpdatedAt = time.Now()
	if n.HealthEndpoint == "" {
		n.HealthEndpoint = "/-/healthy"
	}

	if err := s.repo.Create(ctx, n); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to register collector")
	}

	return n, nil
}

func (s *CollectorService) GetByID(ctx context.Context, id string) (*model.CollectorNode, error) {
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get collector")
	}
	if n == nil {
		return nil, errors.ErrNotFound
	}
	return n, nil
}

func (s *CollectorService) List(ctx context.Context, zoneID, nodeType, status string) ([]*model.CollectorNode, error) {
	return s.repo.List(ctx, zoneID, nodeType, status)
}

func (s *CollectorService) Update(ctx context.Context, id string, updates *model.CollectorNode) (*model.CollectorNode, error) {
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get collector")
	}
	if n == nil {
		return nil, errors.ErrNotFound
	}

	if updates.Name != "" {
		n.Name = updates.Name
	}
	if updates.Address != "" {
		n.Address = updates.Address
	}
	if updates.Status != "" {
		n.Status = updates.Status
	}
	n.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, n); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to update collector")
	}

	return n, nil
}

func (s *CollectorService) Delete(ctx context.Context, id string) error {
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, 50001, 500, "failed to get collector")
	}
	if n == nil {
		return errors.ErrNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return errors.Wrap(err, 50001, 500, "failed to delete collector")
	}

	return nil
}

func (s *CollectorService) Heartbeat(ctx context.Context, id string) error {
	return s.repo.UpdateHeartbeat(ctx, id)
}
