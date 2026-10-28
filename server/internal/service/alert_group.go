package service

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"
	"dcs/pkg/errors"
	"time"

	"github.com/google/uuid"
)

type AlertGroupService struct {
	repo repository.AlertGroupRepo
}

func NewAlertGroupService(repo repository.AlertGroupRepo) *AlertGroupService {
	return &AlertGroupService{repo: repo}
}

func (s *AlertGroupService) Create(ctx context.Context, name, description string, parentID *string, escalationMinutes *int) (*model.AlertGroup, error) {
	if name == "" {
		return nil, errors.Wrap(nil, 40001, 400, "name is required")
	}

	g := &model.AlertGroup{
		GroupID:           "ag-" + uuid.New().String()[:8],
		Name:              name,
		Description:       description,
		ParentID:          parentID,
		EscalationMinutes: escalationMinutes,
		CreatedAt:         time.Now(),
	}

	if err := s.repo.Create(ctx, g); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to create alert group")
	}
	return g, nil
}

func (s *AlertGroupService) GetByID(ctx context.Context, id string) (*model.AlertGroup, error) {
	g, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get alert group")
	}
	if g == nil {
		return nil, errors.ErrNotFound
	}
	return g, nil
}

func (s *AlertGroupService) List(ctx context.Context) ([]*model.AlertGroup, error) {
	return s.repo.List(ctx)
}

func (s *AlertGroupService) Update(ctx context.Context, id, name, description string, parentID *string, escalationMinutes *int) (*model.AlertGroup, error) {
	g, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get alert group")
	}
	if g == nil {
		return nil, errors.ErrNotFound
	}

	if name != "" {
		g.Name = name
	}
	if description != "" {
		g.Description = description
	}
	g.ParentID = parentID
	g.EscalationMinutes = escalationMinutes

	if err := s.repo.Update(ctx, g); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to update alert group")
	}
	return g, nil
}

func (s *AlertGroupService) Delete(ctx context.Context, id string) error {
	g, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, 50001, 500, "failed to get alert group")
	}
	if g == nil {
		return errors.ErrNotFound
	}
	return s.repo.Delete(ctx, id)
}

func (s *AlertGroupService) AddMember(ctx context.Context, groupID, userID, role string) (*model.AlertGroupMember, error) {
	g, err := s.repo.GetByID(ctx, groupID)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get alert group")
	}
	if g == nil {
		return nil, errors.ErrNotFound
	}

	if role == "" {
		role = "member"
	}

	m := &model.AlertGroupMember{
		GroupID:   groupID,
		UserID:    userID,
		Role:      role,
		CreatedAt: time.Now(),
	}

	if err := s.repo.AddMember(ctx, m); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to add member")
	}
	return m, nil
}

func (s *AlertGroupService) RemoveMember(ctx context.Context, groupID, userID string) error {
	return s.repo.RemoveMember(ctx, groupID, userID)
}

func (s *AlertGroupService) ListMembers(ctx context.Context, groupID string) ([]*model.AlertGroupMember, error) {
	return s.repo.ListMembers(ctx, groupID)
}
