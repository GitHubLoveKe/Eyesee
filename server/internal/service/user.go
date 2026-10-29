package service

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"
	"dcs/pkg/errors"
	"time"

	"github.com/google/uuid"
)

type UserService struct {
	repo repository.UserRepo
}

func NewUserService(repo repository.UserRepo) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(ctx context.Context, username, displayName, email, phone, role string) (*model.SysUser, error) {
	if username == "" {
		return nil, errors.Wrap(nil, 40001, 400, "username is required")
	}

	existing, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to check username")
	}
	if existing != nil {
		return nil, errors.Wrap(nil, 40901, 409, "username already exists")
	}

	if role == "" {
		role = "viewer"
	}

	u := &model.SysUser{
		UserID:      "user-" + uuid.New().String()[:8],
		Username:    username,
		DisplayName: displayName,
		Email:       email,
		Phone:       phone,
		Role:        role,
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to create user")
	}
	return u, nil
}

func (s *UserService) GetByID(ctx context.Context, id string) (*model.SysUser, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get user")
	}
	if u == nil {
		return nil, errors.ErrNotFound
	}
	return u, nil
}

func (s *UserService) List(ctx context.Context, role, status, keyword string, page, pageSize int) ([]*model.SysUser, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	filter := repository.UserFilter{
		Role:    role,
		Status:  status,
		Keyword: keyword,
		Offset:  (page - 1) * pageSize,
		Limit:   pageSize,
	}
	return s.repo.List(ctx, filter)
}

func (s *UserService) Update(ctx context.Context, id, displayName, email, phone, role, status string) (*model.SysUser, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get user")
	}
	if u == nil {
		return nil, errors.ErrNotFound
	}

	if displayName != "" {
		u.DisplayName = displayName
	}
	if email != "" {
		u.Email = email
	}
	if phone != "" {
		u.Phone = phone
	}
	if role != "" {
		u.Role = role
	}
	if status != "" {
		u.Status = status
	}
	u.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, u); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to update user")
	}
	return u, nil
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, 50001, 500, "failed to get user")
	}
	if u == nil {
		return errors.ErrNotFound
	}
	return s.repo.Delete(ctx, id)
}
