package postgres

import (
	"context"
	"dcs/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type rcWorkerRepo struct {
	pool *pgxpool.Pool
}

func NewRCWorkerRepo(pool *pgxpool.Pool) *rcWorkerRepo {
	return &rcWorkerRepo{pool: pool}
}

func (r *rcWorkerRepo) Create(ctx context.Context, w *model.RCWorker) error {
	query := `INSERT INTO rc_worker (worker_id, zone_id, name, address, query_address, capacity, owned_rules, status, last_heartbeat, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := r.pool.Exec(ctx, query, w.WorkerID, w.ZoneID, w.Name, w.Address, w.QueryAddress, w.Capacity, w.OwnedRules, w.Status, w.LastHeartbeat, w.CreatedAt, w.UpdatedAt)
	return err
}

func (r *rcWorkerRepo) GetByID(ctx context.Context, id string) (*model.RCWorker, error) {
	w := &model.RCWorker{}
	query := `SELECT worker_id, zone_id, name, address, query_address, capacity, owned_rules, status, last_heartbeat, created_at, updated_at
		FROM rc_worker WHERE worker_id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&w.WorkerID, &w.ZoneID, &w.Name, &w.Address, &w.QueryAddress,
		&w.Capacity, &w.OwnedRules, &w.Status, &w.LastHeartbeat, &w.CreatedAt, &w.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return w, err
}

func (r *rcWorkerRepo) List(ctx context.Context, zoneID string) ([]*model.RCWorker, error) {
	var rows pgx.Rows
	var err error

	if zoneID != "" {
		query := `SELECT worker_id, zone_id, name, address, query_address, capacity, owned_rules, status, last_heartbeat, created_at, updated_at
			FROM rc_worker WHERE zone_id = $1 ORDER BY created_at DESC`
		rows, err = r.pool.Query(ctx, query, zoneID)
	} else {
		query := `SELECT worker_id, zone_id, name, address, query_address, capacity, owned_rules, status, last_heartbeat, created_at, updated_at
			FROM rc_worker ORDER BY created_at DESC`
		rows, err = r.pool.Query(ctx, query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.RCWorker
	for rows.Next() {
		w := &model.RCWorker{}
		if err := rows.Scan(&w.WorkerID, &w.ZoneID, &w.Name, &w.Address, &w.QueryAddress,
			&w.Capacity, &w.OwnedRules, &w.Status, &w.LastHeartbeat, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, w)
	}
	return items, nil
}

func (r *rcWorkerRepo) Update(ctx context.Context, w *model.RCWorker) error {
	query := `UPDATE rc_worker SET zone_id = $2, name = $3, address = $4, query_address = $5,
		capacity = $6, owned_rules = $7, status = $8, last_heartbeat = $9, updated_at = $10
		WHERE worker_id = $1`
	_, err := r.pool.Exec(ctx, query, w.WorkerID, w.ZoneID, w.Name, w.Address, w.QueryAddress,
		w.Capacity, w.OwnedRules, w.Status, w.LastHeartbeat, w.UpdatedAt)
	return err
}

func (r *rcWorkerRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM rc_worker WHERE worker_id = $1`, id)
	return err
}

func (r *rcWorkerRepo) UpdateHeartbeat(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE rc_worker SET last_heartbeat = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE worker_id = $1`, id)
	return err
}
