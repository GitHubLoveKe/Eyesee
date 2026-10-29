package service

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"
	"dcs/pkg/errors"
	"time"

	"github.com/google/uuid"
)

type ZoneService struct {
	repo repository.ZoneRepo
}

func NewZoneService(repo repository.ZoneRepo) *ZoneService {
	return &ZoneService{repo: repo}
}

func (s *ZoneService) Create(ctx context.Context, name, description, collectorType string) (*model.Zone, error) {
	if name == "" {
		return nil, errors.Wrap(nil, 40001, 400, "name is required")
	}
	if collectorType == "" {
		collectorType = "vmagent"
	}

	z := &model.Zone{
		ZoneID:        "zone-" + uuid.New().String()[:8],
		Name:          name,
		Description:   description,
		CollectorType: collectorType,
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.Create(ctx, z); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to create zone")
	}

	return z, nil
}

func (s *ZoneService) GetByID(ctx context.Context, id string) (*model.Zone, error) {
	z, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get zone")
	}
	if z == nil {
		return nil, errors.ErrNotFound
	}
	return z, nil
}

func (s *ZoneService) List(ctx context.Context, status string, page, pageSize int) ([]*model.Zone, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, status, offset, pageSize)
}

func (s *ZoneService) Update(ctx context.Context, id, name, description, collectorType, status string) (*model.Zone, error) {
	z, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get zone")
	}
	if z == nil {
		return nil, errors.ErrNotFound
	}

	if name != "" {
		z.Name = name
	}
	if description != "" {
		z.Description = description
	}
	if collectorType != "" {
		z.CollectorType = collectorType
	}
	if status != "" {
		z.Status = status
	}
	z.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, z); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to update zone")
	}

	return z, nil
}

func (s *ZoneService) Delete(ctx context.Context, id string) error {
	z, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, 50001, 500, "failed to get zone")
	}
	if z == nil {
		return errors.ErrNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return errors.Wrap(err, 50001, 500, "failed to delete zone")
	}

	return nil
}

func (s *ZoneService) CreateSegment(ctx context.Context, zoneID, cidr, name, description string) (*model.NetworkSegment, error) {
	z, err := s.repo.GetByID(ctx, zoneID)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get zone")
	}
	if z == nil {
		return nil, errors.ErrNotFound
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

	if err := s.repo.CreateSegment(ctx, seg); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to create segment")
	}

	return seg, nil
}

func (s *ZoneService) ListSegments(ctx context.Context, zoneID string) ([]*model.NetworkSegment, error) {
	return s.repo.ListSegments(ctx, zoneID)
}

func (s *ZoneService) UpdateSegment(ctx context.Context, segmentID int64, cidr, name, description string) error {
	seg := &model.NetworkSegment{
		SegmentID:   segmentID,
		CIDR:        cidr,
		Name:        name,
		Description: description,
	}

	if err := s.repo.UpdateSegment(ctx, seg); err != nil {
		return errors.Wrap(err, 50001, 500, "failed to update segment")
	}

	return nil
}

func (s *ZoneService) DeleteSegment(ctx context.Context, segmentID int64) error {
	if err := s.repo.DeleteSegment(ctx, segmentID); err != nil {
		return errors.Wrap(err, 50001, 500, "failed to delete segment")
	}
	return nil
}
