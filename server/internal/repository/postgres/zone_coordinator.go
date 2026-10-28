package postgres

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type zoneCoordinatorRepo struct {
	pool *pgxpool.Pool
}

func NewZoneCoordinatorRepo(pool *pgxpool.Pool) repository.ZoneCoordinatorRepo {
	return &zoneCoordinatorRepo{pool: pool}
}

func (r *zoneCoordinatorRepo) Create(ctx context.Context, c *model.ZoneCoordinator) error {
	query := `INSERT INTO zone_coordinator (coordinator_id, zone_id, name, address, gossip_addr, role, managed_nodes, status, last_heartbeat, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := r.pool.Exec(ctx, query, c.CoordinatorID, c.ZoneID, c.Name, c.Address, c.GossipAddr, c.Role, c.ManagedNodes, c.Status, c.LastHeartbeat, c.CreatedAt, c.UpdatedAt)
	return err
}

func (r *zoneCoordinatorRepo) GetByID(ctx context.Context, id string) (*model.ZoneCoordinator, error) {
	c := &model.ZoneCoordinator{}
	query := `SELECT coordinator_id, zone_id, name, address, gossip_addr, role, managed_nodes, status, last_heartbeat, created_at, updated_at
		FROM zone_coordinator WHERE coordinator_id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(&c.CoordinatorID, &c.ZoneID, &c.Name, &c.Address, &c.GossipAddr, &c.Role, &c.ManagedNodes, &c.Status, &c.LastHeartbeat, &c.CreatedAt, &c.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return c, err
}

func (r *zoneCoordinatorRepo) List(ctx context.Context, zoneID string) ([]*model.ZoneCoordinator, error) {
	query := `SELECT coordinator_id, zone_id, name, address, gossip_addr, role, managed_nodes, status, last_heartbeat, created_at, updated_at
		FROM zone_coordinator`
	var args []interface{}
	if zoneID != "" {
		query += " WHERE zone_id = $1"
		args = append(args, zoneID)
	}
	query += " ORDER BY created_at"
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.ZoneCoordinator
	for rows.Next() {
		c := &model.ZoneCoordinator{}
		if err := rows.Scan(&c.CoordinatorID, &c.ZoneID, &c.Name, &c.Address, &c.GossipAddr, &c.Role, &c.ManagedNodes, &c.Status, &c.LastHeartbeat, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, nil
}

func (r *zoneCoordinatorRepo) Update(ctx context.Context, c *model.ZoneCoordinator) error {
	query := `UPDATE zone_coordinator SET zone_id = $2, name = $3, address = $4, gossip_addr = $5, role = $6,
		managed_nodes = $7, status = $8, last_heartbeat = $9, updated_at = $10 WHERE coordinator_id = $1`
	_, err := r.pool.Exec(ctx, query, c.CoordinatorID, c.ZoneID, c.Name, c.Address, c.GossipAddr, c.Role, c.ManagedNodes, c.Status, c.LastHeartbeat, time.Now())
	return err
}

func (r *zoneCoordinatorRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM zone_coordinator WHERE coordinator_id = $1`, id)
	return err
}

func (r *zoneCoordinatorRepo) UpdateHeartbeat(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE zone_coordinator SET last_heartbeat = $2, status = 'active', updated_at = $2 WHERE coordinator_id = $1`, id, time.Now())
	return err
}
