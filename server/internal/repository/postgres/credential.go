package postgres

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type credentialRepo struct {
	pool *pgxpool.Pool
}

func NewCredentialRepo(pool *pgxpool.Pool) repository.CredentialRepo {
	return &credentialRepo{pool: pool}
}

func (r *credentialRepo) Create(ctx context.Context, c *model.Credential) error {
	query := `INSERT INTO credential (credential_id, zone_id, name, cred_type, encrypted_data, version, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.pool.Exec(ctx, query, c.CredentialID, c.ZoneID, c.Name, c.CredType, c.EncryptedData, c.Version, c.ExpiresAt, c.CreatedAt, c.UpdatedAt)
	return err
}

func (r *credentialRepo) GetByID(ctx context.Context, id string) (*model.Credential, error) {
	c := &model.Credential{}
	query := `SELECT credential_id, zone_id, name, cred_type, encrypted_data, version, expires_at, created_at, updated_at
		FROM credential WHERE credential_id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&c.CredentialID, &c.ZoneID, &c.Name, &c.CredType, &c.EncryptedData,
		&c.Version, &c.ExpiresAt, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return c, err
}

func (r *credentialRepo) List(ctx context.Context, zoneID, credType string) ([]*model.Credential, error) {
	where := " WHERE 1=1"
	var args []interface{}
	argIdx := 1

	if zoneID != "" {
		where += " AND zone_id = $" + itoa(argIdx)
		args = append(args, zoneID)
		argIdx++
	}
	if credType != "" {
		where += " AND cred_type = $" + itoa(argIdx)
		args = append(args, credType)
		argIdx++
	}

	query := `SELECT credential_id, zone_id, name, cred_type, '' as encrypted_data, version, expires_at, created_at, updated_at
		FROM credential` + where + ` ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.Credential
	for rows.Next() {
		c := &model.Credential{}
		if err := rows.Scan(&c.CredentialID, &c.ZoneID, &c.Name, &c.CredType, &c.EncryptedData,
			&c.Version, &c.ExpiresAt, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, nil
}

func (r *credentialRepo) Update(ctx context.Context, c *model.Credential) error {
	query := `UPDATE credential SET zone_id = $2, name = $3, cred_type = $4, encrypted_data = $5,
		version = $6, expires_at = $7, updated_at = $8 WHERE credential_id = $1`
	_, err := r.pool.Exec(ctx, query, c.CredentialID, c.ZoneID, c.Name, c.CredType, c.EncryptedData,
		c.Version, c.ExpiresAt, c.UpdatedAt)
	return err
}

func (r *credentialRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM credential WHERE credential_id = $1`, id)
	return err
}

func (r *credentialRepo) HasInstances(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) > 0 FROM instance WHERE credential_id = $1`, id).Scan(&exists)
	return exists, err
}
