package postgres

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type userNotificationTargetRepo struct {
	pool *pgxpool.Pool
}

func NewUserNotificationTargetRepo(pool *pgxpool.Pool) repository.UserNotificationTargetRepo {
	return &userNotificationTargetRepo{pool: pool}
}

func (r *userNotificationTargetRepo) Create(ctx context.Context, t *model.UserNotificationTarget) error {
	query := `INSERT INTO user_notification_target (user_id, channel_type, target_value, is_primary, created_at)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`
	return r.pool.QueryRow(ctx, query, t.UserID, t.ChannelType, t.TargetValue, t.IsPrimary, t.CreatedAt).Scan(&t.ID)
}

func (r *userNotificationTargetRepo) ListByUserID(ctx context.Context, userID string) ([]*model.UserNotificationTarget, error) {
	query := `SELECT id, user_id, channel_type, target_value, is_primary, created_at
		FROM user_notification_target WHERE user_id = $1 ORDER BY is_primary DESC, created_at`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.UserNotificationTarget
	for rows.Next() {
		t := &model.UserNotificationTarget{}
		if err := rows.Scan(&t.ID, &t.UserID, &t.ChannelType, &t.TargetValue, &t.IsPrimary, &t.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	return items, nil
}

func (r *userNotificationTargetRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM user_notification_target WHERE id = $1`, id)
	return err
}

func (r *userNotificationTargetRepo) SetPrimary(ctx context.Context, userID string, id int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `UPDATE user_notification_target SET is_primary = FALSE WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE user_notification_target SET is_primary = TRUE WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
