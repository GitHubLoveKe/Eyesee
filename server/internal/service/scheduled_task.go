package service

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"
	"dcs/pkg/errors"
	"time"

	"github.com/google/uuid"
)

type ScheduledTaskService struct {
	repo repository.ScheduledTaskRepo
}

func NewScheduledTaskService(repo repository.ScheduledTaskRepo) *ScheduledTaskService {
	return &ScheduledTaskService{repo: repo}
}

func (s *ScheduledTaskService) Create(ctx context.Context, req CreateScheduledTaskRequest) (*model.ScheduledTask, error) {
	if req.Name == "" {
		return nil, errors.Wrap(nil, 40001, 400, "name is required")
	}
	if req.TaskType == "" {
		return nil, errors.Wrap(nil, 40001, 400, "task_type is required")
	}
	if req.CronExpr == "" {
		return nil, errors.Wrap(nil, 40001, 400, "cron_expr is required")
	}

	t := &model.ScheduledTask{
		TaskID:        "st-" + uuid.New().String()[:8],
		BusinessSysID: req.BusinessSysID,
		Name:          req.Name,
		TaskType:      req.TaskType,
		CronExpr:      req.CronExpr,
		Config:        req.Config,
		Enabled:       true,
		Status:        "idle",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.Create(ctx, t); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to create scheduled task")
	}
	return t, nil
}

type CreateScheduledTaskRequest struct {
	BusinessSysID *string
	Name          string
	TaskType      string
	CronExpr      string
	Config        string
}

func (s *ScheduledTaskService) GetByID(ctx context.Context, id string) (*model.ScheduledTask, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get scheduled task")
	}
	if t == nil {
		return nil, errors.ErrNotFound
	}
	return t, nil
}

func (s *ScheduledTaskService) List(ctx context.Context, taskType, status string, enabled *bool, keyword string, page, pageSize int) ([]*model.ScheduledTask, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	filter := repository.ScheduledTaskFilter{
		TaskType: taskType,
		Status:   status,
		Enabled:  enabled,
		Keyword:  keyword,
		Offset:   (page - 1) * pageSize,
		Limit:    pageSize,
	}
	return s.repo.List(ctx, filter)
}

func (s *ScheduledTaskService) Update(ctx context.Context, id string, req UpdateScheduledTaskRequest) (*model.ScheduledTask, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get scheduled task")
	}
	if t == nil {
		return nil, errors.ErrNotFound
	}

	if req.Name != "" {
		t.Name = req.Name
	}
	if req.TaskType != "" {
		t.TaskType = req.TaskType
	}
	if req.CronExpr != "" {
		t.CronExpr = req.CronExpr
	}
	if req.Config != "" {
		t.Config = req.Config
	}
	if req.Enabled != nil {
		t.Enabled = *req.Enabled
	}
	t.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, t); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to update scheduled task")
	}
	return t, nil
}

type UpdateScheduledTaskRequest struct {
	BusinessSysID *string
	Name          string
	TaskType      string
	CronExpr      string
	Config        string
	Enabled       *bool
}

func (s *ScheduledTaskService) Delete(ctx context.Context, id string) error {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, 50001, 500, "failed to get scheduled task")
	}
	if t == nil {
		return errors.ErrNotFound
	}
	return s.repo.Delete(ctx, id)
}

func (s *ScheduledTaskService) Enable(ctx context.Context, id string) (*model.ScheduledTask, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get scheduled task")
	}
	if t == nil {
		return nil, errors.ErrNotFound
	}
	t.Enabled = true
	t.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, t); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to enable scheduled task")
	}
	return t, nil
}

func (s *ScheduledTaskService) Disable(ctx context.Context, id string) (*model.ScheduledTask, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get scheduled task")
	}
	if t == nil {
		return nil, errors.ErrNotFound
	}
	t.Enabled = false
	t.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, t); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to disable scheduled task")
	}
	return t, nil
}

func (s *ScheduledTaskService) ListExecutions(ctx context.Context, taskID string, limit int) ([]*model.ScheduledTaskExecution, error) {
	if limit < 1 {
		limit = 20
	}
	return s.repo.ListExecutions(ctx, taskID, limit)
}

func (s *ScheduledTaskService) Trigger(ctx context.Context, taskID string) (*model.ScheduledTaskExecution, error) {
	t, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get scheduled task")
	}
	if t == nil {
		return nil, errors.ErrNotFound
	}

	exec := &model.ScheduledTaskExecution{
		TaskID:    taskID,
		StartedAt: time.Now(),
		Status:    "running",
	}

	if err := s.repo.CreateExecution(ctx, exec); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to create execution")
	}

	now := time.Now()
	t.LastRunAt = &now
	t.UpdatedAt = now
	_ = s.repo.Update(ctx, t)

	return exec, nil
}
