package service

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"
	"dcs/pkg/errors"
	"time"

	"github.com/google/uuid"
)

type CredentialService struct {
	repo repository.CredentialRepo
}

func NewCredentialService(repo repository.CredentialRepo) *CredentialService {
	return &CredentialService{repo: repo}
}

func (s *CredentialService) Create(ctx context.Context, req CreateCredentialRequest) (*model.Credential, error) {
	if req.Name == "" {
		return nil, errors.Wrap(nil, 40001, 400, "name is required")
	}
	if req.ZoneID == "" {
		return nil, errors.Wrap(nil, 40001, 400, "zone_id is required")
	}
	if req.CredType == "" {
		return nil, errors.Wrap(nil, 40001, 400, "cred_type is required")
	}

	c := &model.Credential{
		CredentialID:  "cred-" + uuid.New().String()[:8],
		ZoneID:        req.ZoneID,
		Name:          req.Name,
		CredType:      req.CredType,
		EncryptedData: req.Data,
		Version:       1,
		ExpiresAt:     req.ExpiresAt,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.Create(ctx, c); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to create credential")
	}

	c.EncryptedData = ""
	return c, nil
}

type CreateCredentialRequest struct {
	ZoneID    string
	Name      string
	CredType  string
	Data      string
	ExpiresAt *time.Time
}

func (s *CredentialService) GetByID(ctx context.Context, id string) (*model.Credential, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get credential")
	}
	if c == nil {
		return nil, errors.ErrNotFound
	}
	c.EncryptedData = ""
	return c, nil
}

func (s *CredentialService) List(ctx context.Context, zoneID, credType string) ([]*model.Credential, error) {
	return s.repo.List(ctx, zoneID, credType)
}

func (s *CredentialService) Update(ctx context.Context, id string, req UpdateCredentialRequest) (*model.Credential, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get credential")
	}
	if c == nil {
		return nil, errors.ErrNotFound
	}

	if req.Name != "" {
		c.Name = req.Name
	}
	if req.CredType != "" {
		c.CredType = req.CredType
	}
	if req.Data != "" {
		c.EncryptedData = req.Data
		c.Version++
	}
	c.ExpiresAt = req.ExpiresAt
	c.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, c); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to update credential")
	}
	c.EncryptedData = ""
	return c, nil
}

type UpdateCredentialRequest struct {
	Name      string
	CredType  string
	Data      string
	ExpiresAt *time.Time
}

func (s *CredentialService) Delete(ctx context.Context, id string) error {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, 50001, 500, "failed to get credential")
	}
	if c == nil {
		return errors.ErrNotFound
	}

	hasInst, err := s.repo.HasInstances(ctx, id)
	if err != nil {
		return errors.Wrap(err, 50001, 500, "failed to check instances")
	}
	if hasInst {
		return errors.Wrap(nil, 42202, 422, "credential is referenced by instances")
	}

	return s.repo.Delete(ctx, id)
}
