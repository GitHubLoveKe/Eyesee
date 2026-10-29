package service

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"
	"dcs/pkg/errors"
	"time"

	"github.com/google/uuid"
)

type AlertStrategyService struct {
	repo repository.AlertStrategyRepo
}

func NewAlertStrategyService(repo repository.AlertStrategyRepo) *AlertStrategyService {
	return &AlertStrategyService{repo: repo}
}

func (s *AlertStrategyService) Create(ctx context.Context, req CreateAlertStrategyRequest) (*model.AlertStrategy, error) {
	if req.Name == "" {
		return nil, errors.Wrap(nil, 40001, 400, "name is required")
	}
	if req.MetricID == "" {
		return nil, errors.Wrap(nil, 40001, 400, "metric_id is required")
	}
	if req.ExprTemplate == "" {
		return nil, errors.Wrap(nil, 40001, 400, "expr_template is required")
	}

	severity := req.Severity
	if severity == "" {
		severity = "warning"
	}
	scopeType := req.ScopeType
	if scopeType == "" {
		scopeType = "instance"
	}
	duration := req.Duration
	if duration == "" {
		duration = "5m"
	}

	s2 := &model.AlertStrategy{
		StrategyID:   "as-" + uuid.New().String()[:8],
		Name:         req.Name,
		Description:  req.Description,
		MetricID:     req.MetricID,
		ExprTemplate: req.ExprTemplate,
		Duration:     duration,
		Severity:     severity,
		ScopeType:    scopeType,
		ScopeValue:   req.ScopeValue,
		AlertGroupID: req.AlertGroupID,
		ChannelID:    req.ChannelID,
		Labels:       req.Labels,
		Annotations:  req.Annotations,
		Enabled:      true,
		Status:       "draft",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.Create(ctx, s2); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to create alert strategy")
	}
	return s2, nil
}

type CreateAlertStrategyRequest struct {
	Name         string
	Description  string
	MetricID     string
	ExprTemplate string
	Duration     string
	Severity     string
	ScopeType    string
	ScopeValue   string
	AlertGroupID *string
	ChannelID    *string
	Labels       string
	Annotations  string
}

func (s *AlertStrategyService) GetByID(ctx context.Context, id string) (*model.AlertStrategy, error) {
	s2, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get alert strategy")
	}
	if s2 == nil {
		return nil, errors.ErrNotFound
	}
	return s2, nil
}

func (s *AlertStrategyService) List(ctx context.Context, severity, status string, enabled *bool, keyword string, page, pageSize int) ([]*model.AlertStrategy, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	filter := repository.AlertStrategyFilter{
		Severity: severity,
		Status:   status,
		Enabled:  enabled,
		Keyword:  keyword,
		Offset:   (page - 1) * pageSize,
		Limit:    pageSize,
	}
	return s.repo.List(ctx, filter)
}

func (s *AlertStrategyService) Update(ctx context.Context, id string, req UpdateAlertStrategyRequest) (*model.AlertStrategy, error) {
	s2, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get alert strategy")
	}
	if s2 == nil {
		return nil, errors.ErrNotFound
	}

	if req.Name != "" {
		s2.Name = req.Name
	}
	if req.Description != "" {
		s2.Description = req.Description
	}
	if req.MetricID != "" {
		s2.MetricID = req.MetricID
	}
	if req.ExprTemplate != "" {
		s2.ExprTemplate = req.ExprTemplate
	}
	if req.Duration != "" {
		s2.Duration = req.Duration
	}
	if req.Severity != "" {
		s2.Severity = req.Severity
	}
	if req.ScopeType != "" {
		s2.ScopeType = req.ScopeType
	}
	if req.ScopeValue != "" {
		s2.ScopeValue = req.ScopeValue
	}
	s2.AlertGroupID = req.AlertGroupID
	s2.ChannelID = req.ChannelID
	if req.Labels != "" {
		s2.Labels = req.Labels
	}
	if req.Annotations != "" {
		s2.Annotations = req.Annotations
	}
	if req.Enabled != nil {
		s2.Enabled = *req.Enabled
	}
	if req.Status != "" {
		s2.Status = req.Status
	}
	s2.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, s2); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to update alert strategy")
	}
	return s2, nil
}

type UpdateAlertStrategyRequest struct {
	Name         string
	Description  string
	MetricID     string
	ExprTemplate string
	Duration     string
	Severity     string
	ScopeType    string
	ScopeValue   string
	AlertGroupID *string
	ChannelID    *string
	Labels       string
	Annotations  string
	Enabled      *bool
	Status       string
}

func (s *AlertStrategyService) Delete(ctx context.Context, id string) error {
	s2, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, 50001, 500, "failed to get alert strategy")
	}
	if s2 == nil {
		return errors.ErrNotFound
	}
	return s.repo.Delete(ctx, id)
}

func (s *AlertStrategyService) LinkInstance(ctx context.Context, strategyID, instanceID, zoneID string) (*model.AlertStrategyInstance, error) {
	si := &model.AlertStrategyInstance{
		StrategyID: strategyID,
		InstanceID: instanceID,
		ZoneID:     zoneID,
		CreatedAt:  time.Now(),
	}
	if err := s.repo.LinkInstance(ctx, si); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to link instance")
	}
	return si, nil
}

func (s *AlertStrategyService) UnlinkInstance(ctx context.Context, strategyID, instanceID string) error {
	return s.repo.UnlinkInstance(ctx, strategyID, instanceID)
}

func (s *AlertStrategyService) ListLinkedInstances(ctx context.Context, strategyID string) ([]*model.AlertStrategyInstance, error) {
	return s.repo.ListLinkedInstances(ctx, strategyID)
}
