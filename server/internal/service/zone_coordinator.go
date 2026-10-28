package service

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"
	"dcs/pkg/errors"
	"time"

	"github.com/google/uuid"
)

type ZoneCoordinatorService struct {
	repo repository.ZoneCoordinatorRepo
}

func NewZoneCoordinatorService(repo repository.ZoneCoordinatorRepo) *ZoneCoordinatorService {
	return &ZoneCoordinatorService{repo: repo}
}

func (s *ZoneCoordinatorService) Create(ctx context.Context, req CreateZoneCoordinatorRequest) (*model.ZoneCoordinator, error) {
	if req.Name == "" {
		return nil, errors.Wrap(nil, 40001, 400, "name is required")
	}
	if req.ZoneID == "" {
		return nil, errors.Wrap(nil, 40001, 400, "zone_id is required")
	}

	c := &model.ZoneCoordinator{
		CoordinatorID: "zc-" + uuid.New().String()[:8],
		ZoneID:        req.ZoneID,
		Name:          req.Name,
		Address:       req.Address,
		GossipAddr:    req.GossipAddr,
		Role:          req.Role,
		Status:        "unknown",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if c.Role == "" {
		c.Role = "leader"
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to create coordinator")
	}
	return c, nil
}

type CreateZoneCoordinatorRequest struct {
	ZoneID     string
	Name       string
	Address    string
	GossipAddr string
	Role       string
}

func (s *ZoneCoordinatorService) GetByID(ctx context.Context, id string) (*model.ZoneCoordinator, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get coordinator")
	}
	if c == nil {
		return nil, errors.ErrNotFound
	}
	return c, nil
}

func (s *ZoneCoordinatorService) List(ctx context.Context, zoneID string) ([]*model.ZoneCoordinator, error) {
	return s.repo.List(ctx, zoneID)
}

func (s *ZoneCoordinatorService) Update(ctx context.Context, id string, req UpdateZoneCoordinatorRequest) (*model.ZoneCoordinator, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get coordinator")
	}
	if c == nil {
		return nil, errors.ErrNotFound
	}

	if req.Name != "" {
		c.Name = req.Name
	}
	if req.Address != "" {
		c.Address = req.Address
	}
	if req.GossipAddr != "" {
		c.GossipAddr = req.GossipAddr
	}
	if req.Role != "" {
		c.Role = req.Role
	}
	if req.Status != "" {
		c.Status = req.Status
	}
	c.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, c); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to update coordinator")
	}
	return c, nil
}

type UpdateZoneCoordinatorRequest struct {
	Name       string
	Address    string
	GossipAddr string
	Role       string
	Status     string
}

func (s *ZoneCoordinatorService) Delete(ctx context.Context, id string) error {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, 50001, 500, "failed to get coordinator")
	}
	if c == nil {
		return errors.ErrNotFound
	}
	return s.repo.Delete(ctx, id)
}

func (s *ZoneCoordinatorService) Heartbeat(ctx context.Context, id string) error {
	return s.repo.UpdateHeartbeat(ctx, id)
}
