package postgres

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type businessSystemRepo struct {
	pool *pgxpool.Pool
}

func NewBusinessSystemRepo(pool *pgxpool.Pool) repository.BusinessSystemRepo {
	return &businessSystemRepo{pool: pool}
}

func (r *businessSystemRepo) Create(ctx context.Context, bs *model.BusinessSystem) error {
	query := `INSERT INTO business_system (system_id, name, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.pool.Exec(ctx, query, bs.SystemID, bs.Name, bs.Description, bs.Status, bs.CreatedAt, bs.UpdatedAt)
	return err
}

func (r *businessSystemRepo) GetByID(ctx context.Context, id string) (*model.BusinessSystem, error) {
	bs := &model.BusinessSystem{}
	query := `SELECT system_id, name, description, status, created_at, updated_at
		FROM business_system WHERE system_id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&bs.SystemID, &bs.Name, &bs.Description, &bs.Status, &bs.CreatedAt, &bs.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return bs, err
}

func (r *businessSystemRepo) List(ctx context.Context, keyword string, offset, limit int) ([]*model.BusinessSystem, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM business_system`
	listQuery := `SELECT system_id, name, description, status, created_at, updated_at
		FROM business_system ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	args := []interface{}{}
	if keyword != "" {
		countQuery += ` WHERE name ILIKE $1`
		listQuery = `SELECT system_id, name, description, status, created_at, updated_at
			FROM business_system WHERE name ILIKE $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		args = append(args, "%"+keyword+"%")
	}

	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*model.BusinessSystem
	for rows.Next() {
		bs := &model.BusinessSystem{}
		if err := rows.Scan(&bs.SystemID, &bs.Name, &bs.Description, &bs.Status, &bs.CreatedAt, &bs.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, bs)
	}
	return items, total, nil
}

func (r *businessSystemRepo) Update(ctx context.Context, bs *model.BusinessSystem) error {
	query := `UPDATE business_system SET name = $2, description = $3, status = $4, updated_at = $5
		WHERE system_id = $1`
	_, err := r.pool.Exec(ctx, query, bs.SystemID, bs.Name, bs.Description, bs.Status, bs.UpdatedAt)
	return err
}

func (r *businessSystemRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM business_system WHERE system_id = $1`, id)
	return err
}

func (r *businessSystemRepo) AddMember(ctx context.Context, m *model.BusinessSystemMember) error {
	query := `INSERT INTO business_system_member (system_id, user_id, role, created_at)
		VALUES ($1, $2, $3, $4)`
	_, err := r.pool.Exec(ctx, query, m.SystemID, m.UserID, m.Role, m.CreatedAt)
	return err
}

func (r *businessSystemRepo) RemoveMember(ctx context.Context, systemID, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM business_system_member WHERE system_id = $1 AND user_id = $2`, systemID, userID)
	return err
}

func (r *businessSystemRepo) ListMembers(ctx context.Context, systemID string) ([]*model.BusinessSystemMember, error) {
	query := `SELECT id, system_id, user_id, role, created_at FROM business_system_member WHERE system_id = $1`
	rows, err := r.pool.Query(ctx, query, systemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*model.BusinessSystemMember
	for rows.Next() {
		m := &model.BusinessSystemMember{}
		if err := rows.Scan(&m.ID, &m.SystemID, &m.UserID, &m.Role, &m.CreatedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}

func (r *businessSystemRepo) LinkZone(ctx context.Context, systemID, zoneID string) error {
	query := `INSERT INTO business_system_zone (system_id, zone_id, created_at) VALUES ($1, $2, $3)`
	_, err := r.pool.Exec(ctx, query, systemID, zoneID, time.Now())
	return err
}

func (r *businessSystemRepo) UnlinkZone(ctx context.Context, systemID, zoneID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM business_system_zone WHERE system_id = $1 AND zone_id = $2`, systemID, zoneID)
	return err
}

func (r *businessSystemRepo) UpdateMemberRole(ctx context.Context, systemID, userID, role string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE business_system_member SET role = $3 WHERE system_id = $1 AND user_id = $2`,
		systemID, userID, role)
	return err
}

func (r *businessSystemRepo) ListByUserID(ctx context.Context, userID string) ([]*model.BusinessSystem, error) {
	query := `SELECT bs.system_id, bs.name, bs.description, bs.status, bs.created_at, bs.updated_at
		FROM business_system bs
		INNER JOIN business_system_member bsm ON bs.system_id = bsm.system_id
		WHERE bsm.user_id = $1
		ORDER BY bs.created_at DESC`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.BusinessSystem
	for rows.Next() {
		bs := &model.BusinessSystem{}
		if err := rows.Scan(&bs.SystemID, &bs.Name, &bs.Description, &bs.Status, &bs.CreatedAt, &bs.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, bs)
	}
	return items, nil
}
