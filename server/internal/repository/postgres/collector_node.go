package postgres

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type collectorNodeRepo struct {
	pool *pgxpool.Pool
}

func NewCollectorNodeRepo(pool *pgxpool.Pool) repository.CollectorNodeRepo {
	return &collectorNodeRepo{pool: pool}
}

func (r *collectorNodeRepo) Create(ctx context.Context, n *model.CollectorNode) error {
	query := `INSERT INTO collector_node (node_id, zone_id, name, node_type, address, health_endpoint,
		labels, status, task_count, last_heartbeat, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := r.pool.Exec(ctx, query, n.NodeID, n.ZoneID, n.Name, n.NodeType, n.Address,
		n.HealthEndpoint, n.Labels, n.Status, n.TaskCount, n.LastHeartbeat, n.CreatedAt, n.UpdatedAt)
	return err
}

func (r *collectorNodeRepo) GetByID(ctx context.Context, id string) (*model.CollectorNode, error) {
	n := &model.CollectorNode{}
	query := `SELECT node_id, zone_id, name, node_type, address, health_endpoint, labels, status,
		task_count, last_heartbeat, created_at, updated_at
		FROM collector_node WHERE node_id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&n.NodeID, &n.ZoneID, &n.Name, &n.NodeType, &n.Address, &n.HealthEndpoint,
		&n.Labels, &n.Status, &n.TaskCount, &n.LastHeartbeat, &n.CreatedAt, &n.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return n, err
}

func (r *collectorNodeRepo) List(ctx context.Context, zoneID, nodeType, status string) ([]*model.CollectorNode, error) {
	query := `SELECT node_id, zone_id, name, node_type, address, health_endpoint, labels, status,
		task_count, last_heartbeat, created_at, updated_at
		FROM collector_node WHERE 1=1`
	var args []interface{}
	argIdx := 1

	if zoneID != "" {
		query += " AND zone_id = $" + itoa(argIdx)
		args = append(args, zoneID)
		argIdx++
	}
	if nodeType != "" {
		query += " AND node_type = $" + itoa(argIdx)
		args = append(args, nodeType)
		argIdx++
	}
	if status != "" {
		query += " AND status = $" + itoa(argIdx)
		args = append(args, status)
		argIdx++
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.CollectorNode
	for rows.Next() {
		n := &model.CollectorNode{}
		if err := rows.Scan(&n.NodeID, &n.ZoneID, &n.Name, &n.NodeType, &n.Address,
			&n.HealthEndpoint, &n.Labels, &n.Status, &n.TaskCount,
			&n.LastHeartbeat, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, n)
	}
	return items, nil
}

func (r *collectorNodeRepo) Update(ctx context.Context, n *model.CollectorNode) error {
	query := `UPDATE collector_node SET zone_id = $2, name = $3, node_type = $4, address = $5,
		health_endpoint = $6, labels = $7, status = $8, task_count = $9, updated_at = $10
		WHERE node_id = $1`
	_, err := r.pool.Exec(ctx, query, n.NodeID, n.ZoneID, n.Name, n.NodeType, n.Address,
		n.HealthEndpoint, n.Labels, n.Status, n.TaskCount, n.UpdatedAt)
	return err
}

func (r *collectorNodeRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM collector_node WHERE node_id = $1`, id)
	return err
}

func (r *collectorNodeRepo) UpdateHeartbeat(ctx context.Context, id string) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, `UPDATE collector_node SET last_heartbeat = $2, status = 'healthy', updated_at = $2
		WHERE node_id = $1`, id, now)
	return err
}
