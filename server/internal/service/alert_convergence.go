package service

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"
	"dcs/pkg/errors"
	"time"

	"github.com/google/uuid"
)

type AlertConvergenceService struct {
	eventRepo   repository.AlertEventRepo
	ackRepo     repository.AlertAckRepo
	silenceRepo repository.AlertSilenceRepo
}

func NewAlertConvergenceService(eventRepo repository.AlertEventRepo, ackRepo repository.AlertAckRepo, silenceRepo repository.AlertSilenceRepo) *AlertConvergenceService {
	return &AlertConvergenceService{eventRepo: eventRepo, ackRepo: ackRepo, silenceRepo: silenceRepo}
}

// Alert event methods

func (s *AlertConvergenceService) CreateEvent(ctx context.Context, req CreateAlertEventRequest) (*model.AlertEvent, error) {
	e := &model.AlertEvent{
		AlertName:   req.AlertName,
		ZoneID:      req.ZoneID,
		InstanceID:  req.InstanceID,
		Severity:    req.Severity,
		Status:      strPtr("firing"),
		Labels:      req.Labels,
		Annotations: req.Annotations,
		StartsAt:    req.StartsAt,
		Fingerprint: req.Fingerprint,
		CreatedAt:   time.Now(),
	}
	if err := s.eventRepo.Create(ctx, e); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to create alert event")
	}
	return e, nil
}

type CreateAlertEventRequest struct {
	AlertName   string
	ZoneID      *string
	InstanceID  *string
	Severity    *string
	Labels      string
	Annotations string
	StartsAt    *time.Time
	Fingerprint *string
}

func (s *AlertConvergenceService) GetEvent(ctx context.Context, id int64) (*model.AlertEvent, error) {
	e, err := s.eventRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get alert event")
	}
	if e == nil {
		return nil, errors.ErrNotFound
	}
	return e, nil
}

func (s *AlertConvergenceService) ListEvents(ctx context.Context, filter repository.AlertEventFilter) ([]*model.AlertEvent, int64, error) {
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}
	return s.eventRepo.List(ctx, filter)
}

func (s *AlertConvergenceService) AcknowledgeEvent(ctx context.Context, eventID int64, userID string, note string) (*model.AlertAcknowledgement, error) {
	e, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get alert event")
	}
	if e == nil {
		return nil, errors.ErrNotFound
	}

	ack := &model.AlertAcknowledgement{
		EventID:        eventID,
		UserID:         userID,
		Note:           note,
		AcknowledgedAt: time.Now(),
	}
	if err := s.ackRepo.Create(ctx, ack); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to acknowledge event")
	}
	return ack, nil
}

func (s *AlertConvergenceService) GetStatistics(ctx context.Context, zoneID, period string) (*repository.AlertStatistics, error) {
	return s.eventRepo.Statistics(ctx, zoneID, period)
}

// Silence methods

func (s *AlertConvergenceService) CreateSilence(ctx context.Context, req CreateSilenceRequest) (*model.AlertSilence, error) {
	if req.Matchers == "" {
		return nil, errors.Wrap(nil, 40001, 400, "matchers is required")
	}

	duration := req.Duration
	if duration == "" {
		duration = "1h"
	}

	now := time.Now()
	endsAt := now.Add(parseDuration(duration))

	silence := &model.AlertSilence{
		SilenceID: "sil-" + uuid.New().String()[:8],
		ZoneID:    req.ZoneID,
		Matchers:  req.Matchers,
		StartsAt:  now,
		EndsAt:    endsAt,
		Duration:  strPtr(duration),
		Reason:    req.Reason,
		CreatedBy: req.CreatedBy,
		Status:    "active",
		CreatedAt: now,
	}
	if err := s.silenceRepo.Create(ctx, silence); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to create silence")
	}
	return silence, nil
}

type CreateSilenceRequest struct {
	ZoneID    *string
	Matchers  string
	Duration  string
	Reason    string
	CreatedBy string
}

func (s *AlertConvergenceService) ListSilences(ctx context.Context, zoneID, status string) ([]*model.AlertSilence, error) {
	return s.silenceRepo.List(ctx, zoneID, status)
}

func (s *AlertConvergenceService) RevokeSilence(ctx context.Context, id string) error {
	sil, err := s.silenceRepo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, 50001, 500, "failed to get silence")
	}
	if sil == nil {
		return errors.ErrNotFound
	}
	return s.silenceRepo.Revoke(ctx, id)
}

// helpers
func strPtr(s string) *string { return &s }

func parseDuration(d string) time.Duration {
	if len(d) < 2 {
		return time.Hour
	}
	num := d[:len(d)-1]
	unit := d[len(d)-1]
	var n int
	for _, c := range num {
		n = n*10 + int(c-'0')
	}
	switch unit {
	case 'm':
		return time.Duration(n) * time.Minute
	case 'h':
		return time.Duration(n) * time.Hour
	case 'd':
		return time.Duration(n) * 24 * time.Hour
	default:
		return time.Hour
	}
}
