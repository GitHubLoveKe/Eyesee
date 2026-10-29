package service

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"
	"dcs/pkg/errors"
	"time"

	"github.com/google/uuid"
)

type RCTaskService struct {
	repo repository.RCTaskRepo
}

func NewRCTaskService(repo repository.RCTaskRepo) *RCTaskService {
	return &RCTaskService{repo: repo}
}

func (s *RCTaskService) Create(ctx context.Context, req CreateRCTaskRequest) (*model.RCTask, error) {
	if req.Name == "" {
		return nil, errors.Wrap(nil, 40001, 400, "name is required")
	}
	if req.ZoneID == "" {
		return nil, errors.Wrap(nil, 40001, 400, "zone_id is required")
	}

	t := &model.RCTask{
		TaskID:       "rc-" + uuid.New().String()[:8],
		ZoneID:       req.ZoneID,
		Name:         req.Name,
		Description:  req.Description,
		StrategyID:   req.StrategyID,
		QueryAddress: req.QueryAddress,
		Enabled:      true,
		Status:       "draft",
		RuleGroup:    req.RuleGroup,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.Create(ctx, t); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to create rc task")
	}
	return t, nil
}

type CreateRCTaskRequest struct {
	ZoneID       string
	Name         string
	Description  string
	StrategyID   *string
	QueryAddress *string
	RuleGroup    *string
}

func (s *RCTaskService) GetByID(ctx context.Context, id string) (*model.RCTask, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get rc task")
	}
	if t == nil {
		return nil, errors.ErrNotFound
	}
	return t, nil
}

func (s *RCTaskService) List(ctx context.Context, zoneID string, enabled *bool, keyword string, page, pageSize int) ([]*model.RCTask, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	filter := repository.RCTaskFilter{
		ZoneID:  zoneID,
		Enabled: enabled,
		Keyword: keyword,
		Offset:  (page - 1) * pageSize,
		Limit:   pageSize,
	}
	return s.repo.List(ctx, filter)
}

func (s *RCTaskService) Update(ctx context.Context, id string, req UpdateRCTaskRequest) (*model.RCTask, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get rc task")
	}
	if t == nil {
		return nil, errors.ErrNotFound
	}

	if req.Name != "" {
		t.Name = req.Name
	}
	if req.Description != "" {
		t.Description = req.Description
	}
	if req.QueryAddress != nil {
		t.QueryAddress = req.QueryAddress
	}
	if req.Enabled != nil {
		t.Enabled = *req.Enabled
	}
	if req.Status != "" {
		t.Status = req.Status
	}
	if req.RuleGroup != nil {
		t.RuleGroup = req.RuleGroup
	}
	t.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, t); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to update rc task")
	}
	return t, nil
}

type UpdateRCTaskRequest struct {
	Name         string
	Description  string
	QueryAddress *string
	Enabled      *bool
	Status       string
	RuleGroup    *string
}

func (s *RCTaskService) Delete(ctx context.Context, id string) error {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, 50001, 500, "failed to get rc task")
	}
	if t == nil {
		return errors.ErrNotFound
	}
	return s.repo.Delete(ctx, id)
}

func (s *RCTaskService) CreateRule(ctx context.Context, taskID string, req CreateRCRuleRequest) (*model.RCRule, error) {
	t, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get rc task")
	}
	if t == nil {
		return nil, errors.ErrNotFound
	}

	module := req.Module
	if module == "" {
		module = "metrics"
	}
	severity := req.Severity
	if severity == "" {
		severity = "warning"
	}
	duration := req.Duration
	if duration == "" {
		duration = "5m"
	}

	rule := &model.RCRule{
		RuleID:      "rule-" + uuid.New().String()[:8],
		TaskID:      taskID,
		Module:      module,
		AlertName:   req.AlertName,
		Expr:        req.Expr,
		Duration:    duration,
		Severity:    severity,
		Labels:      req.Labels,
		Annotations: req.Annotations,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.CreateRule(ctx, rule); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to create rule")
	}
	return rule, nil
}

type CreateRCRuleRequest struct {
	Module      string
	AlertName   string
	Expr        string
	Duration    string
	Severity    string
	Labels      string
	Annotations string
}

func (s *RCTaskService) ListRules(ctx context.Context, taskID string) ([]*model.RCRule, error) {
	return s.repo.ListRules(ctx, taskID)
}

func (s *RCTaskService) UpdateRule(ctx context.Context, ruleID string, req UpdateRCRuleRequest) error {
	rule, err := s.repo.GetRuleByID(ctx, ruleID)
	if err != nil {
		return errors.Wrap(err, 50001, 500, "failed to get rule")
	}
	if rule == nil {
		return errors.ErrNotFound
	}

	if req.AlertName != "" {
		rule.AlertName = req.AlertName
	}
	if req.Expr != "" {
		rule.Expr = req.Expr
	}
	if req.Duration != "" {
		rule.Duration = req.Duration
	}
	if req.Severity != "" {
		rule.Severity = req.Severity
	}
	if req.Labels != "" {
		rule.Labels = req.Labels
	}
	if req.Annotations != "" {
		rule.Annotations = req.Annotations
	}

	return s.repo.UpdateRule(ctx, rule)
}

type UpdateRCRuleRequest struct {
	AlertName   string
	Expr        string
	Duration    string
	Severity    string
	Labels      string
	Annotations string
}

func (s *RCTaskService) DeleteRule(ctx context.Context, ruleID string) error {
	return s.repo.DeleteRule(ctx, ruleID)
}
