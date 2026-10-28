package postgres

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type notificationChannelRepo struct {
	pool *pgxpool.Pool
}

func NewNotificationChannelRepo(pool *pgxpool.Pool) repository.NotificationChannelRepo {
	return &notificationChannelRepo{pool: pool}
}

func (r *notificationChannelRepo) Create(ctx context.Context, nc *model.NotificationChannel) error {
	query := `INSERT INTO notification_channel (channel_id, name, channel_type, config, enabled, last_test_at, last_test_result, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.pool.Exec(ctx, query, nc.ChannelID, nc.Name, nc.ChannelType, nc.Config, nc.Enabled, nc.LastTestAt, nc.LastTestResult, nc.CreatedAt, nc.UpdatedAt)
	return err
}

func (r *notificationChannelRepo) GetByID(ctx context.Context, id string) (*model.NotificationChannel, error) {
	nc := &model.NotificationChannel{}
	query := `SELECT channel_id, name, channel_type, config, enabled, last_test_at, last_test_result, created_at, updated_at
		FROM notification_channel WHERE channel_id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&nc.ChannelID, &nc.Name, &nc.ChannelType, &nc.Config, &nc.Enabled,
		&nc.LastTestAt, &nc.LastTestResult, &nc.CreatedAt, &nc.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return nc, err
}

func (r *notificationChannelRepo) List(ctx context.Context, channelType string) ([]*model.NotificationChannel, error) {
	var rows pgx.Rows
	var err error

	if channelType != "" {
		query := `SELECT channel_id, name, channel_type, config, enabled, last_test_at, last_test_result, created_at, updated_at
			FROM notification_channel WHERE channel_type = $1 ORDER BY created_at DESC`
		rows, err = r.pool.Query(ctx, query, channelType)
	} else {
		query := `SELECT channel_id, name, channel_type, config, enabled, last_test_at, last_test_result, created_at, updated_at
			FROM notification_channel ORDER BY created_at DESC`
		rows, err = r.pool.Query(ctx, query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.NotificationChannel
	for rows.Next() {
		nc := &model.NotificationChannel{}
		if err := rows.Scan(&nc.ChannelID, &nc.Name, &nc.ChannelType, &nc.Config, &nc.Enabled,
			&nc.LastTestAt, &nc.LastTestResult, &nc.CreatedAt, &nc.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, nc)
	}
	return items, nil
}

func (r *notificationChannelRepo) Update(ctx context.Context, nc *model.NotificationChannel) error {
	query := `UPDATE notification_channel SET name = $2, channel_type = $3, config = $4, enabled = $5, updated_at = $6
		WHERE channel_id = $1`
	_, err := r.pool.Exec(ctx, query, nc.ChannelID, nc.Name, nc.ChannelType, nc.Config, nc.Enabled, nc.UpdatedAt)
	return err
}

func (r *notificationChannelRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM notification_channel WHERE channel_id = $1`, id)
	return err
}

func (r *notificationChannelRepo) UpdateTestResult(ctx context.Context, id string, result string) error {
	_, err := r.pool.Exec(ctx, `UPDATE notification_channel SET last_test_result = $2, last_test_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE channel_id = $1`, id, result)
	return err
}
