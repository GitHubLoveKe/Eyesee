package postgres

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) repository.UserRepo {
	return &userRepo{pool: pool}
}

func (r *userRepo) Create(ctx context.Context, u *model.SysUser) error {
	query := `INSERT INTO sys_user (user_id, username, display_name, email, phone, role, status, last_login_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.pool.Exec(ctx, query, u.UserID, u.Username, u.DisplayName, u.Email, u.Phone, u.Role, u.Status, u.LastLoginAt, u.CreatedAt, u.UpdatedAt)
	return err
}

func (r *userRepo) GetByID(ctx context.Context, id string) (*model.SysUser, error) {
	u := &model.SysUser{}
	query := `SELECT user_id, username, COALESCE(display_name,''), COALESCE(email,''), COALESCE(phone,''), role, status, last_login_at, created_at, updated_at
		FROM sys_user WHERE user_id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.UserID, &u.Username, &u.DisplayName, &u.Email, &u.Phone,
		&u.Role, &u.Status, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (*model.SysUser, error) {
	u := &model.SysUser{}
	query := `SELECT user_id, username, COALESCE(display_name,''), COALESCE(email,''), COALESCE(phone,''), role, status, last_login_at, created_at, updated_at
		FROM sys_user WHERE username = $1`
	err := r.pool.QueryRow(ctx, query, username).Scan(
		&u.UserID, &u.Username, &u.DisplayName, &u.Email, &u.Phone,
		&u.Role, &u.Status, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (r *userRepo) List(ctx context.Context, f repository.UserFilter) ([]*model.SysUser, int64, error) {
	where := " WHERE 1=1"
	var args []interface{}
	argIdx := 1

	if f.Role != "" {
		where += " AND role = $" + itoa(argIdx)
		args = append(args, f.Role)
		argIdx++
	}
	if f.Status != "" {
		where += " AND status = $" + itoa(argIdx)
		args = append(args, f.Status)
		argIdx++
	}
	if f.Keyword != "" {
		where += " AND (username ILIKE $" + itoa(argIdx) + " OR display_name ILIKE $" + itoa(argIdx) + ")"
		args = append(args, "%"+f.Keyword+"%")
		argIdx++
	}

	var total int64
	countQuery := "SELECT COUNT(*) FROM sys_user" + where
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQuery := `SELECT user_id, username, COALESCE(display_name,''), COALESCE(email,''), COALESCE(phone,''), role, status, last_login_at, created_at, updated_at
		FROM sys_user` + where + ` ORDER BY created_at DESC LIMIT $` + itoa(argIdx) + ` OFFSET $` + itoa(argIdx+1)
	args = append(args, f.Limit, f.Offset)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*model.SysUser
	for rows.Next() {
		u := &model.SysUser{}
		if err := rows.Scan(&u.UserID, &u.Username, &u.DisplayName, &u.Email, &u.Phone,
			&u.Role, &u.Status, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, u)
	}
	return items, total, nil
}

func (r *userRepo) Update(ctx context.Context, u *model.SysUser) error {
	query := `UPDATE sys_user SET display_name = $2, email = $3, phone = $4, role = $5, status = $6, updated_at = $7
		WHERE user_id = $1`
	_, err := r.pool.Exec(ctx, query, u.UserID, u.DisplayName, u.Email, u.Phone, u.Role, u.Status, u.UpdatedAt)
	return err
}

func (r *userRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sys_user WHERE user_id = $1`, id)
	return err
}

func (r *userRepo) UpdateLastLogin(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE sys_user SET last_login_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE user_id = $1`, id)
	return err
}
