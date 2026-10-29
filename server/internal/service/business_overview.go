package service

import (
	"context"
	"dcs/internal/repository"
	"dcs/pkg/errors"
)

type BusinessOverviewService struct {
	repo repository.BusinessOverviewRepo
}

func NewBusinessOverviewService(repo repository.BusinessOverviewRepo) *BusinessOverviewService {
	return &BusinessOverviewService{repo: repo}
}

func (s *BusinessOverviewService) GetOverview(ctx context.Context, systemID string) (*repository.BusinessOverview, error) {
	if systemID == "" {
		return nil, errors.Wrap(nil, 40001, 400, "system_id is required")
	}
	overview, err := s.repo.GetOverview(ctx, systemID)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get business overview")
	}
	return overview, nil
}
