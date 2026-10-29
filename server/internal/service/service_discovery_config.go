package service

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"
	"dcs/pkg/errors"
	"time"

	"github.com/google/uuid"
)

type ServiceDiscoveryConfigService struct {
	repo repository.ServiceDiscoveryConfigRepo
}

func NewServiceDiscoveryConfigService(repo repository.ServiceDiscoveryConfigRepo) *ServiceDiscoveryConfigService {
	return &ServiceDiscoveryConfigService{repo: repo}
}

func (s *ServiceDiscoveryConfigService) Create(ctx context.Context, req CreateSDConfigRequest) (*model.ServiceDiscovery, error) {
	if req.Name == "" {
		return nil, errors.Wrap(nil, 40001, 400, "name is required")
	}
	if req.ZoneID == "" {
		return nil, errors.Wrap(nil, 40001, 400, "zone_id is required")
	}

	protocol := req.Protocol
	if protocol == "" {
		protocol = "http_sd"
	}

	sd := &model.ServiceDiscovery{
		DiscoveryID:     "sd-" + uuid.New().String()[:8],
		ZoneID:          req.ZoneID,
		Name:            req.Name,
		Protocol:        protocol,
		Config:          req.Config,
		RefreshInterval: req.RefreshInterval,
		Enabled:         true,
		Status:          "active",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if sd.RefreshInterval == "" {
		sd.RefreshInterval = "30s"
	}
	if err := s.repo.Create(ctx, sd); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to create service discovery config")
	}
	return sd, nil
}

type CreateSDConfigRequest struct {
	ZoneID          string
	Name            string
	Protocol        string
	Config          string
	RefreshInterval string
}

func (s *ServiceDiscoveryConfigService) GetByID(ctx context.Context, id string) (*model.ServiceDiscovery, error) {
	sd, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get service discovery config")
	}
	if sd == nil {
		return nil, errors.ErrNotFound
	}
	return sd, nil
}

func (s *ServiceDiscoveryConfigService) List(ctx context.Context, zoneID string, page, pageSize int) ([]*model.ServiceDiscovery, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.List(ctx, zoneID)
}

func (s *ServiceDiscoveryConfigService) Update(ctx context.Context, id string, req UpdateSDConfigRequest) (*model.ServiceDiscovery, error) {
	sd, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get service discovery config")
	}
	if sd == nil {
		return nil, errors.ErrNotFound
	}

	if req.Name != "" {
		sd.Name = req.Name
	}
	if req.Protocol != "" {
		sd.Protocol = req.Protocol
	}
	if req.Config != "" {
		sd.Config = req.Config
	}
	if req.RefreshInterval != "" {
		sd.RefreshInterval = req.RefreshInterval
	}
	if req.Enabled != nil {
		sd.Enabled = *req.Enabled
	}
	sd.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, sd); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to update service discovery config")
	}
	return sd, nil
}

type UpdateSDConfigRequest struct {
	Name            string
	Protocol        string
	Config          string
	RefreshInterval string
	Enabled         *bool
}

func (s *ServiceDiscoveryConfigService) Delete(ctx context.Context, id string) error {
	sd, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, 50001, 500, "failed to get service discovery config")
	}
	if sd == nil {
		return errors.ErrNotFound
	}
	return s.repo.Delete(ctx, id)
}
