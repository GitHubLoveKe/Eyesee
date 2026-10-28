package postgres

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type alertGroupRepo struct {
	pool *pgxpool.Pool
}

func NewAlertGroupRepo(pool *pgxpool.Pool) repository.AlertGroupRepo {
	return &alertGroupRepo{pool: pool}
}

func (r *alertGroupRepo) Create(ctx context.Context, g *model.AlertGroup) error {
	query := `INSERT INTO alert_group (group_id, name, description, parent_id, escalation_minutes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.pool.Exec(ctx, query, g.GroupID, g.Name, g.Description, g.ParentID, g.EscalationMinutes, g.CreatedAt)
	return err
}

func (r *alertGroupRepo) GetByID(ctx context.Context, id string) (*model.AlertGroup, error) {
	g := &model.AlertGroup{}
	query := `SELECT group_id, name, description, parent_id, escalation_minutes, created_at
		FROM alert_group WHERE group_id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&g.GroupID, &g.Name, &g.Description, &g.ParentID, &g.EscalationMinutes, &g.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return g, err
}

func (r *alertGroupRepo) List(ctx context.Context) ([]*model.AlertGroup, error) {
	query := `SELECT group_id, name, description, parent_id, escalation_minutes, created_at
		FROM alert_group ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.AlertGroup
	for rows.Next() {
		g := &model.AlertGroup{}
		if err := rows.Scan(&g.GroupID, &g.Name, &g.Description, &g.ParentID, &g.EscalationMinutes, &g.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, g)
	}
	return items, nil
}

func (r *alertGroupRepo) Update(ctx context.Context, g *model.AlertGroup) error {
	query := `UPDATE alert_group SET name = $2, description = $3, parent_id = $4, escalation_minutes = $5
		WHERE group_id = $1`
	_, err := r.pool.Exec(ctx, query, g.GroupID, g.Name, g.Description, g.ParentID, g.EscalationMinutes)
	return err
}

func (r *alertGroupRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM alert_group WHERE group_id = $1`, id)
	return err
}

func (r *alertGroupRepo) AddMember(ctx context.Context, m *model.AlertGroupMember) error {
	query := `INSERT INTO alert_group_member (group_id, user_id, role, created_at)
		VALUES ($1, $2, $3, $4) RETURNING id`
	return r.pool.QueryRow(ctx, query, m.GroupID, m.UserID, m.Role, m.CreatedAt).Scan(&m.ID)
}

func (r *alertGroupRepo) RemoveMember(ctx context.Context, groupID, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM alert_group_member WHERE group_id = $1 AND user_id = $2`, groupID, userID)
	return err
}

func (r *alertGroupRepo) ListMembers(ctx context.Context, groupID string) ([]*model.AlertGroupMember, error) {
	query := `SELECT id, group_id, user_id, role, created_at
		FROM alert_group_member WHERE group_id = $1 ORDER BY created_at`
	rows, err := r.pool.Query(ctx, query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.AlertGroupMember
	for rows.Next() {
		m := &model.AlertGroupMember{}
		if err := rows.Scan(&m.ID, &m.GroupID, &m.UserID, &m.Role, &m.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return items, nil
}
