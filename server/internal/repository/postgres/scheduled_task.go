package postgres

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type scheduledTaskRepo struct {
	pool *pgxpool.Pool
}

func NewScheduledTaskRepo(pool *pgxpool.Pool) repository.ScheduledTaskRepo {
	return &scheduledTaskRepo{pool: pool}
}

func (r *scheduledTaskRepo) Create(ctx context.Context, t *model.ScheduledTask) error {
	query := `INSERT INTO scheduled_task (task_id, business_system_id, name, task_type, cron_expr, config, enabled, last_run_at, next_run_at, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := r.pool.Exec(ctx, query, t.TaskID, t.BusinessSysID, t.Name, t.TaskType, t.CronExpr, t.Config, t.Enabled, t.LastRunAt, t.NextRunAt, t.Status, t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *scheduledTaskRepo) GetByID(ctx context.Context, id string) (*model.ScheduledTask, error) {
	t := &model.ScheduledTask{}
	query := `SELECT task_id, business_system_id, name, task_type, cron_expr, config, enabled, last_run_at, next_run_at, status, created_at, updated_at
		FROM scheduled_task WHERE task_id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&t.TaskID, &t.BusinessSysID, &t.Name, &t.TaskType, &t.CronExpr, &t.Config,
		&t.Enabled, &t.LastRunAt, &t.NextRunAt, &t.Status, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return t, err
}

func (r *scheduledTaskRepo) List(ctx context.Context, f repository.ScheduledTaskFilter) ([]*model.ScheduledTask, int64, error) {
	where := " WHERE 1=1"
	var args []interface{}
	argIdx := 1

	if f.TaskType != "" {
		where += " AND task_type = $" + itoa(argIdx)
		args = append(args, f.TaskType)
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
		where += " AND name ILIKE $" + itoa(argIdx)
		args = append(args, "%"+f.Keyword+"%")
		argIdx++
	}

	var total int64
	countQuery := "SELECT COUNT(*) FROM scheduled_task" + where
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQuery := `SELECT task_id, business_system_id, name, task_type, cron_expr, config, enabled, last_run_at, next_run_at, status, created_at, updated_at
		FROM scheduled_task` + where + ` ORDER BY created_at DESC LIMIT $` + itoa(argIdx) + ` OFFSET $` + itoa(argIdx+1)
	args = append(args, f.Limit, f.Offset)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*model.ScheduledTask
	for rows.Next() {
		t := &model.ScheduledTask{}
		if err := rows.Scan(&t.TaskID, &t.BusinessSysID, &t.Name, &t.TaskType, &t.CronExpr, &t.Config,
			&t.Enabled, &t.LastRunAt, &t.NextRunAt, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, t)
	}
	return items, total, nil
}

func (r *scheduledTaskRepo) Update(ctx context.Context, t *model.ScheduledTask) error {
	query := `UPDATE scheduled_task SET business_system_id = $2, name = $3, task_type = $4, cron_expr = $5,
		config = $6, enabled = $7, last_run_at = $8, next_run_at = $9, status = $10, updated_at = $11
		WHERE task_id = $1`
	_, err := r.pool.Exec(ctx, query, t.TaskID, t.BusinessSysID, t.Name, t.TaskType, t.CronExpr,
		t.Config, t.Enabled, t.LastRunAt, t.NextRunAt, t.Status, t.UpdatedAt)
	return err
}

func (r *scheduledTaskRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM scheduled_task WHERE task_id = $1`, id)
	return err
}

func (r *scheduledTaskRepo) CreateExecution(ctx context.Context, e *model.ScheduledTaskExecution) error {
	query := `INSERT INTO scheduled_task_execution (task_id, started_at, completed_at, duration_ms, status, error_message)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING execution_id`
	return r.pool.QueryRow(ctx, query, e.TaskID, e.StartedAt, e.CompletedAt, e.DurationMs, e.Status, e.ErrorMessage).Scan(&e.ExecutionID)
}

func (r *scheduledTaskRepo) UpdateExecution(ctx context.Context, e *model.ScheduledTaskExecution) error {
	query := `UPDATE scheduled_task_execution SET completed_at = $2, duration_ms = $3, status = $4, error_message = $5
		WHERE execution_id = $1`
	_, err := r.pool.Exec(ctx, query, e.ExecutionID, e.CompletedAt, e.DurationMs, e.Status, e.ErrorMessage)
	return err
}

func (r *scheduledTaskRepo) ListExecutions(ctx context.Context, taskID string, limit int) ([]*model.ScheduledTaskExecution, error) {
	query := `SELECT execution_id, task_id, started_at, completed_at, duration_ms, status, error_message
		FROM scheduled_task_execution WHERE task_id = $1 ORDER BY started_at DESC LIMIT $2`
	rows, err := r.pool.Query(ctx, query, taskID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.ScheduledTaskExecution
	for rows.Next() {
		e := &model.ScheduledTaskExecution{}
		if err := rows.Scan(&e.ExecutionID, &e.TaskID, &e.StartedAt, &e.CompletedAt, &e.DurationMs, &e.Status, &e.ErrorMessage); err != nil {
			return nil, err
		}
		items = append(items, e)
	}
	return items, nil
}
