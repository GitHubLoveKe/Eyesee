package service

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"
	"dcs/pkg/errors"
	"strings"
	"time"
)

type NetworkSegmentService struct {
	repo repository.NetworkSegmentRepo
}

func NewNetworkSegmentService(repo repository.NetworkSegmentRepo) *NetworkSegmentService {
	return &NetworkSegmentService{repo: repo}
}

func (s *NetworkSegmentService) Create(ctx context.Context, zoneID, cidr, name, description string) (*model.NetworkSegment, error) {
	if cidr == "" {
		return nil, errors.Wrap(nil, 40001, 400, "CIDR is required")
	}
	if !strings.Contains(cidr, "/") {
		return nil, errors.Wrap(nil, 40002, 400, "CIDR must contain '/'")
	}
	if name == "" {
		return nil, errors.Wrap(nil, 40001, 400, "name is required")
	}

	exists, err := s.repo.CheckCIDROverlap(ctx, zoneID, cidr, nil)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to check CIDR overlap")
	}
	if exists {
		return nil, errors.Wrap(nil, 40901, 409, "CIDR already exists in this zone")
	}

	seg := &model.NetworkSegment{
		ZoneID:      zoneID,
		CIDR:        cidr,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, seg); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to create network segment")
	}

	return seg, nil
}

func (s *NetworkSegmentService) GetByID(ctx context.Context, id int64) (*model.NetworkSegment, error) {
	seg, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get network segment")
	}
	if seg == nil {
		return nil, errors.ErrNotFound
	}
	return seg, nil
}

func (s *NetworkSegmentService) ListByZoneID(ctx context.Context, zoneID string) ([]*model.NetworkSegment, error) {
	return s.repo.ListByZoneID(ctx, zoneID)
}

func (s *NetworkSegmentService) Update(ctx context.Context, id int64, cidr, name, description string) (*model.NetworkSegment, error) {
	seg, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get network segment")
	}
	if seg == nil {
		return nil, errors.ErrNotFound
	}

	if cidr != "" {
		if !strings.Contains(cidr, "/") {
			return nil, errors.Wrap(nil, 40002, 400, "CIDR must contain '/'")
		}
		exists, err := s.repo.CheckCIDROverlap(ctx, seg.ZoneID, cidr, &id)
		if err != nil {
			return nil, errors.Wrap(err, 50001, 500, "failed to check CIDR overlap")
		}
		if exists {
			return nil, errors.Wrap(nil, 40901, 409, "CIDR already exists in this zone")
		}
		seg.CIDR = cidr
	}
	if name != "" {
		seg.Name = name
	}
	seg.Description = description

	if err := s.repo.Update(ctx, seg); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to update network segment")
	}

	return seg, nil
}

func (s *NetworkSegmentService) Delete(ctx context.Context, id int64) error {
	seg, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, 50001, 500, "failed to get network segment")
	}
	if seg == nil {
		return errors.ErrNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return errors.Wrap(err, 50001, 500, "failed to delete network segment")
	}

	return nil
}
