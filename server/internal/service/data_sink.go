package service

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"
	"dcs/pkg/errors"
	"time"

	"github.com/google/uuid"
)

type DataSinkService struct {
	repo repository.DataSinkRepo
}

func NewDataSinkService(repo repository.DataSinkRepo) *DataSinkService {
	return &DataSinkService{repo: repo}
}

func (s *DataSinkService) Create(ctx context.Context, ds *model.DataSink) (*model.DataSink, error) {
	if ds.Name == "" {
		return nil, errors.Wrap(nil, 40001, 400, "name is required")
	}
	if ds.ZoneID == "" {
		return nil, errors.Wrap(nil, 40001, 400, "zone_id is required")
	}
	if ds.SinkType == "" {
		return nil, errors.Wrap(nil, 40001, 400, "sink_type is required")
	}
	if ds.Protocol == "" {
		return nil, errors.Wrap(nil, 40001, 400, "protocol is required")
	}

	ds.SinkID = "sink-" + uuid.New().String()[:8]
	ds.Status = "active"
	ds.CreatedAt = time.Now()
	ds.UpdatedAt = time.Now()

	if err := s.repo.Create(ctx, ds); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to create data sink")
	}

	return ds, nil
}

func (s *DataSinkService) GetByID(ctx context.Context, id string) (*model.DataSink, error) {
	ds, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get data sink")
	}
	if ds == nil {
		return nil, errors.ErrNotFound
	}
	return ds, nil
}

func (s *DataSinkService) List(ctx context.Context, zoneID, sinkType string) ([]*model.DataSink, error) {
	return s.repo.List(ctx, zoneID, sinkType)
}

func (s *DataSinkService) Update(ctx context.Context, id string, updates *model.DataSink) (*model.DataSink, error) {
	ds, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get data sink")
	}
	if ds == nil {
		return nil, errors.ErrNotFound
	}

	if updates.Name != "" {
		ds.Name = updates.Name
	}
	if updates.SinkType != "" {
		ds.SinkType = updates.SinkType
	}
	if updates.Protocol != "" {
		ds.Protocol = updates.Protocol
	}
	if updates.Config != "" {
		ds.Config = updates.Config
	}
	if updates.Status != "" {
		ds.Status = updates.Status
	}
	ds.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, ds); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to update data sink")
	}

	return ds, nil
}

func (s *DataSinkService) Delete(ctx context.Context, id string) error {
	ds, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, 50001, 500, "failed to get data sink")
	}
	if ds == nil {
		return errors.ErrNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return errors.Wrap(err, 50001, 500, "failed to delete data sink")
	}

	return nil
}

func (s *DataSinkService) UpdateHealthResult(ctx context.Context, id string, checkAt time.Time, result string) error {
	return s.repo.UpdateHealthResult(ctx, id, checkAt, result)
}
