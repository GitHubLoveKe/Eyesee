package postgres

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type collectionTaskRepo struct {
	pool *pgxpool.Pool
}

func NewCollectionTaskRepo(pool *pgxpool.Pool) repository.CollectionTaskRepo {
	return &collectionTaskRepo{pool: pool}
}

func (r *collectionTaskRepo) Create(ctx context.Context, ct *model.CollectionTask) error {
	query := `INSERT INTO collection_task (task_id, instance_id, type_id, module, node_id, data_sink_id,
		name, config, enabled, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := r.pool.Exec(ctx, query, ct.TaskID, ct.InstanceID, ct.TypeID, ct.Module, ct.NodeID,
		ct.DataSinkID, ct.Name, ct.Config, ct.Enabled, ct.Status, ct.CreatedAt, ct.UpdatedAt)
	return err
}

func (r *collectionTaskRepo) GetByID(ctx context.Context, id string) (*model.CollectionTask, error) {
	ct := &model.CollectionTask{}
	query := `SELECT task_id, instance_id, type_id, module, node_id, data_sink_id, name, config,
		enabled, status, last_scrape_at, last_error, created_at, updated_at
		FROM collection_task WHERE task_id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&ct.TaskID, &ct.InstanceID, &ct.TypeID, &ct.Module, &ct.NodeID, &ct.DataSinkID,
		&ct.Name, &ct.Config, &ct.Enabled, &ct.Status, &ct.LastScrape, &ct.LastError,
		&ct.CreatedAt, &ct.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return ct, err
}

func (r *collectionTaskRepo) List(ctx context.Context, filter repository.CollectionTaskFilter) ([]*model.CollectionTask, int64, error) {
	where := " WHERE 1=1"
	var args []interface{}
	argIdx := 1

	if filter.InstanceID != "" {
		where += " AND instance_id = $" + itoa(argIdx)
		args = append(args, filter.InstanceID)
		argIdx++
	}
	if filter.TypeID != "" {
		where += " AND type_id = $" + itoa(argIdx)
		args = append(args, filter.TypeID)
		argIdx++
	}
	if filter.NodeID != "" {
		where += " AND node_id = $" + itoa(argIdx)
		args = append(args, filter.NodeID)
		argIdx++
	}
	if filter.Status != "" {
		where += " AND status = $" + itoa(argIdx)
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.Enabled != nil {
		where += " AND enabled = $" + itoa(argIdx)
		args = append(args, *filter.Enabled)
		argIdx++
	}
	if filter.Keyword != "" {
		where += " AND name ILIKE $" + itoa(argIdx)
		args = append(args, "%"+filter.Keyword+"%")
		argIdx++
	}

	var total int64
	countQuery := "SELECT COUNT(*) FROM collection_task" + where
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQuery := `SELECT task_id, instance_id, type_id, module, node_id, data_sink_id, name, config,
		enabled, status, last_scrape_at, last_error, created_at, updated_at
		FROM collection_task` + where + ` ORDER BY created_at DESC LIMIT $` + itoa(argIdx) + ` OFFSET $` + itoa(argIdx+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*model.CollectionTask
	for rows.Next() {
		ct := &model.CollectionTask{}
		if err := rows.Scan(&ct.TaskID, &ct.InstanceID, &ct.TypeID, &ct.Module, &ct.NodeID,
			&ct.DataSinkID, &ct.Name, &ct.Config, &ct.Enabled, &ct.Status,
			&ct.LastScrape, &ct.LastError, &ct.CreatedAt, &ct.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, ct)
	}
	return items, total, nil
}

func (r *collectionTaskRepo) Update(ctx context.Context, ct *model.CollectionTask) error {
	query := `UPDATE collection_task SET instance_id = $2, type_id = $3, module = $4, node_id = $5,
		data_sink_id = $6, name = $7, config = $8, enabled = $9, status = $10, updated_at = $11
		WHERE task_id = $1`
	_, err := r.pool.Exec(ctx, query, ct.TaskID, ct.InstanceID, ct.TypeID, ct.Module, ct.NodeID,
		ct.DataSinkID, ct.Name, ct.Config, ct.Enabled, ct.Status, ct.UpdatedAt)
	return err
}

func (r *collectionTaskRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM collection_task WHERE task_id = $1`, id)
	return err
}

func (r *collectionTaskRepo) CountByInstance(ctx context.Context, instanceID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM collection_task WHERE instance_id = $1`, instanceID).Scan(&count)
	return count, err
}
