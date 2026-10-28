package postgres

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type metricDefRepo struct {
	pool *pgxpool.Pool
}

func NewMetricDefinitionRepo(pool *pgxpool.Pool) repository.MetricDefinitionRepo {
	return &metricDefRepo{pool: pool}
}

func (r *metricDefRepo) Create(ctx context.Context, m *model.MetricDefinition) error {
	if m.MetricID == "" {
		m.MetricID = "metric-" + uuid.New().String()[:8]
	}
	query := `INSERT INTO metric_definition (metric_id, name, category, metric_type, help, unit, promql_template, variables, labels, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := r.pool.Exec(ctx, query, m.MetricID, m.Name, m.Category, m.MetricType, m.Help, m.Unit, m.PromqlTemplate, m.Variables, m.Labels, m.CreatedAt, m.UpdatedAt)
	return err
}

func (r *metricDefRepo) GetByID(ctx context.Context, id string) (*model.MetricDefinition, error) {
	m := &model.MetricDefinition{}
	query := `SELECT metric_id, name, category, metric_type, help, unit, promql_template, variables, labels, created_at, updated_at
		FROM metric_definition WHERE metric_id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&m.MetricID, &m.Name, &m.Category, &m.MetricType, &m.Help, &m.Unit,
		&m.PromqlTemplate, &m.Variables, &m.Labels, &m.CreatedAt, &m.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return m, err
}

func (r *metricDefRepo) List(ctx context.Context, f repository.MetricFilter) ([]*model.MetricDefinition, int64, error) {
	where := " WHERE 1=1"
	var args []interface{}
	argIdx := 1

	if f.Category != "" {
		where += " AND category = $" + itoa(argIdx)
		args = append(args, f.Category)
		argIdx++
	}
	if f.MetricType != "" {
		where += " AND metric_type = $" + itoa(argIdx)
		args = append(args, f.MetricType)
		argIdx++
	}
	if f.Keyword != "" {
		where += " AND name ILIKE $" + itoa(argIdx)
		args = append(args, "%"+f.Keyword+"%")
		argIdx++
	}

	var total int64
	countQuery := "SELECT COUNT(*) FROM metric_definition" + where
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQuery := `SELECT metric_id, name, category, metric_type, help, unit, promql_template, variables, labels, created_at, updated_at
		FROM metric_definition` + where + ` ORDER BY created_at DESC LIMIT $` + itoa(argIdx) + ` OFFSET $` + itoa(argIdx+1)
	args = append(args, f.Limit, f.Offset)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*model.MetricDefinition
	for rows.Next() {
		m := &model.MetricDefinition{}
		if err := rows.Scan(&m.MetricID, &m.Name, &m.Category, &m.MetricType, &m.Help, &m.Unit,
			&m.PromqlTemplate, &m.Variables, &m.Labels, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	return items, total, nil
}

func (r *metricDefRepo) Update(ctx context.Context, m *model.MetricDefinition) error {
	query := `UPDATE metric_definition SET name = $2, category = $3, metric_type = $4, help = $5, unit = $6,
		promql_template = $7, variables = $8, labels = $9, updated_at = $10 WHERE metric_id = $1`
	_, err := r.pool.Exec(ctx, query, m.MetricID, m.Name, m.Category, m.MetricType, m.Help, m.Unit,
		m.PromqlTemplate, m.Variables, m.Labels, m.UpdatedAt)
	return err
}

func (r *metricDefRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM metric_definition WHERE metric_id = $1`, id)
	return err
}

func (r *metricDefRepo) CreateQuery(ctx context.Context, q *model.CommonQuery) error {
	if q.QueryID == "" {
		q.QueryID = "cq-" + uuid.New().String()[:8]
	}
	query := `INSERT INTO common_query (query_id, name, description, metric_id, promql, variables, category, is_builtin, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.pool.Exec(ctx, query, q.QueryID, q.Name, q.Description, q.MetricID, q.Promql, q.Variables, q.Category, q.IsBuiltin, q.CreatedAt)
	return err
}

func (r *metricDefRepo) ListQueries(ctx context.Context, metricID string) ([]*model.CommonQuery, error) {
	query := `SELECT query_id, name, description, metric_id, promql, variables, category, is_builtin, created_at
		FROM common_query WHERE metric_id = $1 ORDER BY created_at`
	rows, err := r.pool.Query(ctx, query, metricID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.CommonQuery
	for rows.Next() {
		q := &model.CommonQuery{}
		if err := rows.Scan(&q.QueryID, &q.Name, &q.Description, &q.MetricID, &q.Promql, &q.Variables, &q.Category, &q.IsBuiltin, &q.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, q)
	}
	return items, nil
}

func (r *metricDefRepo) DeleteQuery(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM common_query WHERE query_id = $1`, id)
	return err
}
