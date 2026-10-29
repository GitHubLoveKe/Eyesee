package postgres

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type alertSilenceRepo struct {
	pool *pgxpool.Pool
}

func NewAlertSilenceRepo(pool *pgxpool.Pool) repository.AlertSilenceRepo {
	return &alertSilenceRepo{pool: pool}
}

func (r *alertSilenceRepo) Create(ctx context.Context, s *model.AlertSilence) error {
	query := `INSERT INTO alert_silence (silence_id, zone_id, matchers, starts_at, ends_at, duration, reason, created_by, status, engine_silence_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := r.pool.Exec(ctx, query, s.SilenceID, s.ZoneID, s.Matchers, s.StartsAt, s.EndsAt, s.Duration, s.Reason, s.CreatedBy, s.Status, s.EngineSilenceID, s.CreatedAt)
	return err
}

func (r *alertSilenceRepo) GetByID(ctx context.Context, id string) (*model.AlertSilence, error) {
	s := &model.AlertSilence{}
	query := `SELECT silence_id, zone_id, matchers, starts_at, ends_at, duration, reason, created_by, status, engine_silence_id, created_at
		FROM alert_silence WHERE silence_id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(&s.SilenceID, &s.ZoneID, &s.Matchers, &s.StartsAt, &s.EndsAt, &s.Duration, &s.Reason, &s.CreatedBy, &s.Status, &s.EngineSilenceID, &s.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return s, err
}

func (r *alertSilenceRepo) List(ctx context.Context, zoneID, status string) ([]*model.AlertSilence, error) {
	where := " WHERE 1=1"
	var args []interface{}
	argIdx := 1

	if zoneID != "" {
		where += " AND zone_id = $" + itoa(argIdx)
		args = append(args, zoneID)
		argIdx++
	}
	if status != "" {
		where += " AND status = $" + itoa(argIdx)
		args = append(args, status)
		argIdx++
	}

	query := `SELECT silence_id, zone_id, matchers, starts_at, ends_at, duration, reason, created_by, status, engine_silence_id, created_at
		FROM alert_silence` + where + ` ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.AlertSilence
	for rows.Next() {
		s := &model.AlertSilence{}
		if err := rows.Scan(&s.SilenceID, &s.ZoneID, &s.Matchers, &s.StartsAt, &s.EndsAt, &s.Duration, &s.Reason, &s.CreatedBy, &s.Status, &s.EngineSilenceID, &s.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	return items, nil
}

func (r *alertSilenceRepo) Revoke(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE alert_silence SET status = 'revoked' WHERE silence_id = $1`, id)
	return err
}
