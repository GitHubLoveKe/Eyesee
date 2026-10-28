package postgres

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type serviceDiscoveryConfigRepo struct {
	pool *pgxpool.Pool
}

func NewServiceDiscoveryConfigRepo(pool *pgxpool.Pool) repository.ServiceDiscoveryConfigRepo {
	return &serviceDiscoveryConfigRepo{pool: pool}
}

func (r *serviceDiscoveryConfigRepo) Create(ctx context.Context, sd *model.ServiceDiscovery) error {
	query := `INSERT INTO service_discovery (discovery_id, zone_id, name, protocol, config, refresh_interval, enabled, status, last_refresh_at, discovered_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := r.pool.Exec(ctx, query, sd.DiscoveryID, sd.ZoneID, sd.Name, sd.Protocol, sd.Config, sd.RefreshInterval, sd.Enabled, sd.Status, sd.LastRefreshAt, sd.DiscoveredCount, sd.CreatedAt, sd.UpdatedAt)
	return err
}

func (r *serviceDiscoveryConfigRepo) GetByID(ctx context.Context, id string) (*model.ServiceDiscovery, error) {
	sd := &model.ServiceDiscovery{}
	query := `SELECT discovery_id, zone_id, name, protocol, config, refresh_interval, enabled, status, last_refresh_at, discovered_count, created_at, updated_at
		FROM service_discovery WHERE discovery_id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(&sd.DiscoveryID, &sd.ZoneID, &sd.Name, &sd.Protocol, &sd.Config, &sd.RefreshInterval, &sd.Enabled, &sd.Status, &sd.LastRefreshAt, &sd.DiscoveredCount, &sd.CreatedAt, &sd.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return sd, err
}

func (r *serviceDiscoveryConfigRepo) List(ctx context.Context, zoneID string) ([]*model.ServiceDiscovery, int64, error) {
	where := " WHERE 1=1"
	var args []interface{}
	argIdx := 1

	if zoneID != "" {
		where += " AND zone_id = $" + itoa(argIdx)
		args = append(args, zoneID)
		argIdx++
	}

	var total int64
	r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM service_discovery"+where, args...).Scan(&total)

	query := `SELECT discovery_id, zone_id, name, protocol, config, refresh_interval, enabled, status, last_refresh_at, discovered_count, created_at, updated_at
		FROM service_discovery` + where + ` ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*model.ServiceDiscovery
	for rows.Next() {
		sd := &model.ServiceDiscovery{}
		if err := rows.Scan(&sd.DiscoveryID, &sd.ZoneID, &sd.Name, &sd.Protocol, &sd.Config, &sd.RefreshInterval, &sd.Enabled, &sd.Status, &sd.LastRefreshAt, &sd.DiscoveredCount, &sd.CreatedAt, &sd.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, sd)
	}
	return items, total, nil
}

func (r *serviceDiscoveryConfigRepo) Update(ctx context.Context, sd *model.ServiceDiscovery) error {
	query := `UPDATE service_discovery SET zone_id = $2, name = $3, protocol = $4, config = $5, refresh_interval = $6,
		enabled = $7, status = $8, last_refresh_at = $9, discovered_count = $10, updated_at = $11 WHERE discovery_id = $1`
	_, err := r.pool.Exec(ctx, query, sd.DiscoveryID, sd.ZoneID, sd.Name, sd.Protocol, sd.Config, sd.RefreshInterval, sd.Enabled, sd.Status, sd.LastRefreshAt, sd.DiscoveredCount, time.Now())
	return err
}

func (r *serviceDiscoveryConfigRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM service_discovery WHERE discovery_id = $1`, id)
	return err
}
