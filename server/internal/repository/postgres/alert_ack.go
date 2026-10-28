package postgres

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type alertAckRepo struct {
	pool *pgxpool.Pool
}

func NewAlertAckRepo(pool *pgxpool.Pool) repository.AlertAckRepo {
	return &alertAckRepo{pool: pool}
}

func (r *alertAckRepo) Create(ctx context.Context, ack *model.AlertAcknowledgement) error {
	query := `INSERT INTO alert_acknowledgement (event_id, user_id, note, acknowledged_at)
		VALUES ($1, $2, $3, $4) RETURNING id`
	return r.pool.QueryRow(ctx, query, ack.EventID, ack.UserID, ack.Note, ack.AcknowledgedAt).Scan(&ack.ID)
}

func (r *alertAckRepo) GetByEventID(ctx context.Context, eventID int64) (*model.AlertAcknowledgement, error) {
	ack := &model.AlertAcknowledgement{}
	query := `SELECT id, event_id, user_id, note, acknowledged_at FROM alert_acknowledgement WHERE event_id = $1 ORDER BY acknowledged_at DESC LIMIT 1`
	err := r.pool.QueryRow(ctx, query, eventID).Scan(&ack.ID, &ack.EventID, &ack.UserID, &ack.Note, &ack.AcknowledgedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return ack, err
}
