package postgres

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type taskTypeRepo struct {
	pool *pgxpool.Pool
}

func NewTaskTypeRepo(pool *pgxpool.Pool) repository.TaskTypeRepo {
	return &taskTypeRepo{pool: pool}
}

func (r *taskTypeRepo) Create(ctx context.Context, tt *model.TaskType) error {
	query := `INSERT INTO task_type (type_id, name, description, source_type, collector, form_schema, detail_panels, declared_metrics, config_template, is_builtin, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	_, err := r.pool.Exec(ctx, query, tt.TypeID, tt.Name, tt.Description, tt.SourceType, tt.Collector,
		tt.FormSchema, tt.DetailPanels, tt.DeclaredMetrics, tt.ConfigTemplate, tt.IsBuiltin, tt.Version,
		tt.CreatedAt, tt.UpdatedAt)
	return err
}

func (r *taskTypeRepo) GetByID(ctx context.Context, id string) (*model.TaskType, error) {
	tt := &model.TaskType{}
	query := `SELECT type_id, name, description, source_type, collector, form_schema, detail_panels, declared_metrics, config_template, is_builtin, version, created_at, updated_at
		FROM task_type WHERE type_id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&tt.TypeID, &tt.Name, &tt.Description, &tt.SourceType, &tt.Collector,
		&tt.FormSchema, &tt.DetailPanels, &tt.DeclaredMetrics, &tt.ConfigTemplate,
		&tt.IsBuiltin, &tt.Version, &tt.CreatedAt, &tt.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return tt, err
}

func (r *taskTypeRepo) List(ctx context.Context) ([]*model.TaskType, error) {
	query := `SELECT type_id, name, description, source_type, collector, form_schema, detail_panels, declared_metrics, config_template, is_builtin, version, created_at, updated_at
		FROM task_type ORDER BY name`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.TaskType
	for rows.Next() {
		tt := &model.TaskType{}
		if err := rows.Scan(&tt.TypeID, &tt.Name, &tt.Description, &tt.SourceType, &tt.Collector,
			&tt.FormSchema, &tt.DetailPanels, &tt.DeclaredMetrics, &tt.ConfigTemplate,
			&tt.IsBuiltin, &tt.Version, &tt.CreatedAt, &tt.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, tt)
	}
	return items, nil
}

func (r *taskTypeRepo) Update(ctx context.Context, tt *model.TaskType) error {
	query := `UPDATE task_type SET name = $2, description = $3, source_type = $4, collector = $5,
		form_schema = $6, detail_panels = $7, declared_metrics = $8, config_template = $9,
		is_builtin = $10, version = $11, updated_at = $12 WHERE type_id = $1`
	_, err := r.pool.Exec(ctx, query, tt.TypeID, tt.Name, tt.Description, tt.SourceType, tt.Collector,
		tt.FormSchema, tt.DetailPanels, tt.DeclaredMetrics, tt.ConfigTemplate,
		tt.IsBuiltin, tt.Version, tt.UpdatedAt)
	return err
}

func (r *taskTypeRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM task_type WHERE type_id = $1`, id)
	return err
}
