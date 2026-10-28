package postgres

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type resourceTypeMetricRepo struct {
	pool *pgxpool.Pool
}

func NewResourceTypeMetricRepo(pool *pgxpool.Pool) repository.ResourceTypeMetricRepo {
	return &resourceTypeMetricRepo{pool: pool}
}

func (r *resourceTypeMetricRepo) Link(ctx context.Context, typeID, metricID string) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO resource_type_metric (type_id, metric_id, created_at) VALUES ($1, $2, $3) ON CONFLICT (type_id, metric_id) DO NOTHING`, typeID, metricID, time.Now())
	return err
}

func (r *resourceTypeMetricRepo) Unlink(ctx context.Context, typeID, metricID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM resource_type_metric WHERE type_id = $1 AND metric_id = $2`, typeID, metricID)
	return err
}

func (r *resourceTypeMetricRepo) ListByType(ctx context.Context, typeID string) ([]*model.ResourceTypeMetric, error) {
	query := `SELECT id, type_id, metric_id, created_at FROM resource_type_metric WHERE type_id = $1 ORDER BY created_at`
	rows, err := r.pool.Query(ctx, query, typeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.ResourceTypeMetric
	for rows.Next() {
		m := &model.ResourceTypeMetric{}
		if err := rows.Scan(&m.ID, &m.TypeID, &m.MetricID, &m.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return items, nil
}
