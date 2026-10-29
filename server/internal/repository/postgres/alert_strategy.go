package postgres

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type alertStrategyRepo struct {
	pool *pgxpool.Pool
}

func NewAlertStrategyRepo(pool *pgxpool.Pool) repository.AlertStrategyRepo {
	return &alertStrategyRepo{pool: pool}
}

func (r *alertStrategyRepo) Create(ctx context.Context, s *model.AlertStrategy) error {
	query := `INSERT INTO alert_strategy (strategy_id, name, description, metric_id, expr_template, duration,
		severity, scope_type, scope_value, alert_group_id, channel_id, labels, annotations, enabled, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`
	_, err := r.pool.Exec(ctx, query, s.StrategyID, s.Name, s.Description, s.MetricID, s.ExprTemplate,
		s.Duration, s.Severity, s.ScopeType, s.ScopeValue, s.AlertGroupID, s.ChannelID,
		s.Labels, s.Annotations, s.Enabled, s.Status, s.CreatedAt, s.UpdatedAt)
	return err
}

func (r *alertStrategyRepo) GetByID(ctx context.Context, id string) (*model.AlertStrategy, error) {
	s := &model.AlertStrategy{}
	query := `SELECT strategy_id, name, description, metric_id, expr_template, duration,
		severity, scope_type, scope_value, alert_group_id, channel_id, labels, annotations, enabled, status, created_at, updated_at
		FROM alert_strategy WHERE strategy_id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.StrategyID, &s.Name, &s.Description, &s.MetricID, &s.ExprTemplate,
		&s.Duration, &s.Severity, &s.ScopeType, &s.ScopeValue, &s.AlertGroupID, &s.ChannelID,
		&s.Labels, &s.Annotations, &s.Enabled, &s.Status, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return s, err
}

func (r *alertStrategyRepo) List(ctx context.Context, f repository.AlertStrategyFilter) ([]*model.AlertStrategy, int64, error) {
	where := " WHERE 1=1"
	var args []interface{}
	argIdx := 1

	if f.Severity != "" {
		where += " AND severity = $" + itoa(argIdx)
		args = append(args, f.Severity)
		argIdx++
	}
	if f.Status != "" {
		where += " AND status = $" + itoa(argIdx)
		args = append(args, f.Status)
		argIdx++
	}
	if f.Enabled != nil {
		where += " AND enabled = $" + itoa(argIdx)
		args = append(args, *f.Enabled)
		argIdx++
	}
	if f.Keyword != "" {
		where += " AND (name ILIKE $" + itoa(argIdx) + " OR description ILIKE $" + itoa(argIdx) + ")"
		args = append(args, "%"+f.Keyword+"%")
		argIdx++
	}

	var total int64
	countQuery := "SELECT COUNT(*) FROM alert_strategy" + where
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQuery := `SELECT strategy_id, name, description, metric_id, expr_template, duration,
		severity, scope_type, scope_value, alert_group_id, channel_id, labels, annotations, enabled, status, created_at, updated_at
		FROM alert_strategy` + where + ` ORDER BY created_at DESC LIMIT $` + itoa(argIdx) + ` OFFSET $` + itoa(argIdx+1)
	args = append(args, f.Limit, f.Offset)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*model.AlertStrategy
	for rows.Next() {
		s := &model.AlertStrategy{}
		if err := rows.Scan(&s.StrategyID, &s.Name, &s.Description, &s.MetricID, &s.ExprTemplate,
			&s.Duration, &s.Severity, &s.ScopeType, &s.ScopeValue, &s.AlertGroupID, &s.ChannelID,
			&s.Labels, &s.Annotations, &s.Enabled, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, nil
}

func (r *alertStrategyRepo) Update(ctx context.Context, s *model.AlertStrategy) error {
	query := `UPDATE alert_strategy SET name = $2, description = $3, metric_id = $4, expr_template = $5,
		duration = $6, severity = $7, scope_type = $8, scope_value = $9, alert_group_id = $10, channel_id = $11,
		labels = $12, annotations = $13, enabled = $14, status = $15, updated_at = $16
		WHERE strategy_id = $1`
	_, err := r.pool.Exec(ctx, query, s.StrategyID, s.Name, s.Description, s.MetricID, s.ExprTemplate,
		s.Duration, s.Severity, s.ScopeType, s.ScopeValue, s.AlertGroupID, s.ChannelID,
		s.Labels, s.Annotations, s.Enabled, s.Status, s.UpdatedAt)
	return err
}

func (r *alertStrategyRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM alert_strategy WHERE strategy_id = $1`, id)
	return err
}

func (r *alertStrategyRepo) LinkInstance(ctx context.Context, si *model.AlertStrategyInstance) error {
	query := `INSERT INTO alert_strategy_instance (strategy_id, instance_id, zone_id, created_at)
		VALUES ($1, $2, $3, $4) RETURNING id`
	return r.pool.QueryRow(ctx, query, si.StrategyID, si.InstanceID, si.ZoneID, si.CreatedAt).Scan(&si.ID)
}

func (r *alertStrategyRepo) UnlinkInstance(ctx context.Context, strategyID, instanceID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM alert_strategy_instance WHERE strategy_id = $1 AND instance_id = $2`, strategyID, instanceID)
	return err
}

func (r *alertStrategyRepo) ListLinkedInstances(ctx context.Context, strategyID string) ([]*model.AlertStrategyInstance, error) {
	query := `SELECT id, strategy_id, instance_id, zone_id, created_at
		FROM alert_strategy_instance WHERE strategy_id = $1 ORDER BY created_at`
	rows, err := r.pool.Query(ctx, query, strategyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.AlertStrategyInstance
	for rows.Next() {
		si := &model.AlertStrategyInstance{}
		if err := rows.Scan(&si.ID, &si.StrategyID, &si.InstanceID, &si.ZoneID, &si.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, si)
	}
	return items, nil
}
