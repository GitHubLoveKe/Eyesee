package postgres

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userAuthSourceRepo struct {
	pool *pgxpool.Pool
}

func NewUserAuthSourceRepo(pool *pgxpool.Pool) repository.UserAuthSourceRepo {
	return &userAuthSourceRepo{pool: pool}
}

func (r *userAuthSourceRepo) GetByUserIDAndType(ctx context.Context, userID string, sourceType string) (*model.UserAuthSource, error) {
	src := &model.UserAuthSource{}
	query := `SELECT id, user_id, source_type, COALESCE(external_id,''), COALESCE(password_hash,''), COALESCE(extra,''), created_at
		FROM user_auth_source WHERE user_id = $1 AND source_type = $2`
	err := r.pool.QueryRow(ctx, query, userID, sourceType).Scan(
		&src.ID, &src.UserID, &src.SourceType, &src.ExternalID,
		&src.PasswordHash, &src.Extra, &src.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return src, err
}

func (r *userAuthSourceRepo) Create(ctx context.Context, src *model.UserAuthSource) error {
	query := `INSERT INTO user_auth_source (user_id, source_type, external_id, password_hash, extra, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.pool.Exec(ctx, query, src.UserID, src.SourceType, src.ExternalID, src.PasswordHash, src.Extra, src.CreatedAt)
	return err
}

func (r *userAuthSourceRepo) UpdatePassword(ctx context.Context, userID string, sourceType string, passwordHash string) error {
	query := `UPDATE user_auth_source SET password_hash = $3 WHERE user_id = $1 AND source_type = $2`
	_, err := r.pool.Exec(ctx, query, userID, sourceType, passwordHash)
	return err
}
