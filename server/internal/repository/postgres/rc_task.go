package postgres

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type rcTaskRepo struct {
	pool *pgxpool.Pool
}

func NewRCTaskRepo(pool *pgxpool.Pool) repository.RCTaskRepo {
	return &rcTaskRepo{pool: pool}
}

func (r *rcTaskRepo) Create(ctx context.Context, t *model.RCTask) error {
	if t.TaskID == "" {
		t.TaskID = "rc-" + uuid.New().String()[:8]
	}
	query := `INSERT INTO rc_task (task_id, zone_id, name, description, strategy_id, query_address, enabled, status, rule_group, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := r.pool.Exec(ctx, query, t.TaskID, t.ZoneID, t.Name, t.Description, t.StrategyID, t.QueryAddress, t.Enabled, t.Status, t.RuleGroup, t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *rcTaskRepo) GetByID(ctx context.Context, id string) (*model.RCTask, error) {
	t := &model.RCTask{}
	query := `SELECT task_id, zone_id, name, description, strategy_id, query_address, enabled, status, rule_group, created_at, updated_at
		FROM rc_task WHERE task_id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&t.TaskID, &t.ZoneID, &t.Name, &t.Description, &t.StrategyID, &t.QueryAddress,
		&t.Enabled, &t.Status, &t.RuleGroup, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return t, err
}

func (r *rcTaskRepo) List(ctx context.Context, f repository.RCTaskFilter) ([]*model.RCTask, int64, error) {
	where := " WHERE 1=1"
	var args []interface{}
	argIdx := 1

	if f.ZoneID != "" {
		where += " AND zone_id = $" + itoa(argIdx)
		args = append(args, f.ZoneID)
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
	countQuery := "SELECT COUNT(*) FROM rc_task" + where
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQuery := `SELECT task_id, zone_id, name, description, strategy_id, query_address, enabled, status, rule_group, created_at, updated_at
		FROM rc_task` + where + ` ORDER BY created_at DESC LIMIT $` + itoa(argIdx) + ` OFFSET $` + itoa(argIdx+1)
	args = append(args, f.Limit, f.Offset)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*model.RCTask
	for rows.Next() {
		t := &model.RCTask{}
		if err := rows.Scan(&t.TaskID, &t.ZoneID, &t.Name, &t.Description, &t.StrategyID, &t.QueryAddress,
			&t.Enabled, &t.Status, &t.RuleGroup, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, t)
	}
	return items, total, nil
}

func (r *rcTaskRepo) Update(ctx context.Context, t *model.RCTask) error {
	query := `UPDATE rc_task SET zone_id = $2, name = $3, description = $4, strategy_id = $5, query_address = $6,
		enabled = $7, status = $8, rule_group = $9, updated_at = $10 WHERE task_id = $1`
	_, err := r.pool.Exec(ctx, query, t.TaskID, t.ZoneID, t.Name, t.Description, t.StrategyID, t.QueryAddress,
		t.Enabled, t.Status, t.RuleGroup, t.UpdatedAt)
	return err
}

func (r *rcTaskRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM rc_task WHERE task_id = $1`, id)
	return err
}

func (r *rcTaskRepo) CreateRule(ctx context.Context, rule *model.RCRule) error {
	if rule.RuleID == "" {
		rule.RuleID = "rule-" + uuid.New().String()[:8]
	}
	query := `INSERT INTO rc_rule (rule_id, task_id, module, alert_name, expr, duration, severity, labels, annotations, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.pool.Exec(ctx, query, rule.RuleID, rule.TaskID, rule.Module, rule.AlertName, rule.Expr,
		rule.Duration, rule.Severity, rule.Labels, rule.Annotations, rule.CreatedAt)
	return err
}

func (r *rcTaskRepo) GetRuleByID(ctx context.Context, id string) (*model.RCRule, error) {
	rule := &model.RCRule{}
	query := `SELECT rule_id, task_id, module, alert_name, expr, duration, severity, labels, annotations, created_at
		FROM rc_rule WHERE rule_id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&rule.RuleID, &rule.TaskID, &rule.Module, &rule.AlertName, &rule.Expr,
		&rule.Duration, &rule.Severity, &rule.Labels, &rule.Annotations, &rule.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return rule, err
}

func (r *rcTaskRepo) ListRules(ctx context.Context, taskID string) ([]*model.RCRule, error) {
	query := `SELECT rule_id, task_id, module, alert_name, expr, duration, severity, labels, annotations, created_at
		FROM rc_rule WHERE task_id = $1 ORDER BY created_at`
	rows, err := r.pool.Query(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.RCRule
	for rows.Next() {
		rule := &model.RCRule{}
		if err := rows.Scan(&rule.RuleID, &rule.TaskID, &rule.Module, &rule.AlertName, &rule.Expr,
			&rule.Duration, &rule.Severity, &rule.Labels, &rule.Annotations, &rule.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, rule)
	}
	return items, nil
}

func (r *rcTaskRepo) UpdateRule(ctx context.Context, rule *model.RCRule) error {
	query := `UPDATE rc_rule SET module = $2, alert_name = $3, expr = $4, duration = $5, severity = $6, labels = $7, annotations = $8
		WHERE rule_id = $1`
	_, err := r.pool.Exec(ctx, query, rule.RuleID, rule.Module, rule.AlertName, rule.Expr,
		rule.Duration, rule.Severity, rule.Labels, rule.Annotations)
	return err
}

func (r *rcTaskRepo) DeleteRule(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM rc_rule WHERE rule_id = $1`, id)
	return err
}
