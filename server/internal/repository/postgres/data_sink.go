package postgres

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type dataSinkRepo struct {
	pool *pgxpool.Pool
}

func NewDataSinkRepo(pool *pgxpool.Pool) repository.DataSinkRepo {
	return &dataSinkRepo{pool: pool}
}

func (r *dataSinkRepo) Create(ctx context.Context, ds *model.DataSink) error {
	query := `INSERT INTO data_sink (sink_id, zone_id, name, sink_type, protocol, config, is_default, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.pool.Exec(ctx, query, ds.SinkID, ds.ZoneID, ds.Name, ds.SinkType, ds.Protocol,
		ds.Config, ds.IsDefault, ds.Status, ds.CreatedAt, ds.UpdatedAt)
	return err
}

func (r *dataSinkRepo) GetByID(ctx context.Context, id string) (*model.DataSink, error) {
	ds := &model.DataSink{}
	query := `SELECT sink_id, zone_id, name, sink_type, protocol, config, is_default, status, last_check_at, created_at, updated_at
		FROM data_sink WHERE sink_id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&ds.SinkID, &ds.ZoneID, &ds.Name, &ds.SinkType, &ds.Protocol,
		&ds.Config, &ds.IsDefault, &ds.Status, &ds.LastCheckAt, &ds.CreatedAt, &ds.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return ds, err
}

func (r *dataSinkRepo) List(ctx context.Context, zoneID, sinkType string) ([]*model.DataSink, error) {
	query := `SELECT sink_id, zone_id, name, sink_type, protocol, config, is_default, status, last_check_at, created_at, updated_at
		FROM data_sink WHERE 1=1`
	var args []interface{}
	argIdx := 1

	if zoneID != "" {
		query += " AND zone_id = $" + itoa(argIdx)
		args = append(args, zoneID)
		argIdx++
	}
	if sinkType != "" {
		query += " AND sink_type = $" + itoa(argIdx)
		args = append(args, sinkType)
		argIdx++
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.DataSink
	for rows.Next() {
		ds := &model.DataSink{}
		if err := rows.Scan(&ds.SinkID, &ds.ZoneID, &ds.Name, &ds.SinkType, &ds.Protocol,
			&ds.Config, &ds.IsDefault, &ds.Status, &ds.LastCheckAt, &ds.CreatedAt, &ds.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, ds)
	}
	return items, nil
}

func (r *dataSinkRepo) Update(ctx context.Context, ds *model.DataSink) error {
	query := `UPDATE data_sink SET name = $2, sink_type = $3, protocol = $4, config = $5,
		is_default = $6, status = $7, updated_at = $8 WHERE sink_id = $1`
	_, err := r.pool.Exec(ctx, query, ds.SinkID, ds.Name, ds.SinkType, ds.Protocol,
		ds.Config, ds.IsDefault, ds.Status, ds.UpdatedAt)
	return err
}

func (r *dataSinkRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM data_sink WHERE sink_id = $1`, id)
	return err
}
