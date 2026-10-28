package postgres

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type instanceRepo struct {
	pool *pgxpool.Pool
}

func NewInstanceRepo(pool *pgxpool.Pool) repository.InstanceRepo {
	return &instanceRepo{pool: pool}
}

func (r *instanceRepo) Create(ctx context.Context, inst *model.Instance) error {
	query := `INSERT INTO instance (instance_id, zone_id, resource_type_id, business_system_id, name, host, port,
		credential_id, labels, extra, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	_, err := r.pool.Exec(ctx, query, inst.InstanceID, inst.ZoneID, inst.ResourceTypeID,
		inst.BusinessSysID, inst.Name, inst.Host, inst.Port, inst.CredentialID,
		inst.Labels, inst.Extra, inst.Status, inst.CreatedAt, inst.UpdatedAt)
	return err
}

func (r *instanceRepo) GetByID(ctx context.Context, id string) (*model.Instance, error) {
	inst := &model.Instance{}
	query := `SELECT instance_id, zone_id, resource_type_id, business_system_id, name, host, port,
		credential_id, labels, extra, status, last_test_at, last_test_result, created_at, updated_at
		FROM instance WHERE instance_id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&inst.InstanceID, &inst.ZoneID, &inst.ResourceTypeID, &inst.BusinessSysID,
		&inst.Name, &inst.Host, &inst.Port, &inst.CredentialID,
		&inst.Labels, &inst.Extra, &inst.Status, &inst.LastTestAt,
		&inst.LastTestResult, &inst.CreatedAt, &inst.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return inst, err
}

func (r *instanceRepo) List(ctx context.Context, filter repository.InstanceFilter) ([]*model.Instance, int64, error) {
	where := " WHERE 1=1"
	var args []interface{}
	argIdx := 1

	if filter.ZoneID != "" {
		where += " AND zone_id = $" + itoa(argIdx)
		args = append(args, filter.ZoneID)
		argIdx++
	}
	if filter.ResourceTypeID != "" {
		where += " AND resource_type_id = $" + itoa(argIdx)
		args = append(args, filter.ResourceTypeID)
		argIdx++
	}
	if filter.BusinessSysID != "" {
		where += " AND business_system_id = $" + itoa(argIdx)
		args = append(args, filter.BusinessSysID)
		argIdx++
	}
	if filter.Status != "" {
		where += " AND status = $" + itoa(argIdx)
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.Keyword != "" {
		where += " AND (name ILIKE $" + itoa(argIdx) + " OR host ILIKE $" + itoa(argIdx) + ")"
		args = append(args, "%"+filter.Keyword+"%")
		argIdx++
	}

	var total int64
	countQuery := "SELECT COUNT(*) FROM instance" + where
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQuery := `SELECT instance_id, zone_id, resource_type_id, business_system_id, name, host, port,
		credential_id, labels, extra, status, last_test_at, last_test_result, created_at, updated_at
		FROM instance` + where + ` ORDER BY created_at DESC LIMIT $` + itoa(argIdx) + ` OFFSET $` + itoa(argIdx+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*model.Instance
	for rows.Next() {
		inst := &model.Instance{}
		if err := rows.Scan(&inst.InstanceID, &inst.ZoneID, &inst.ResourceTypeID, &inst.BusinessSysID,
			&inst.Name, &inst.Host, &inst.Port, &inst.CredentialID,
			&inst.Labels, &inst.Extra, &inst.Status, &inst.LastTestAt,
			&inst.LastTestResult, &inst.CreatedAt, &inst.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, inst)
	}
	return items, total, nil
}

func (r *instanceRepo) Update(ctx context.Context, inst *model.Instance) error {
	query := `UPDATE instance SET zone_id = $2, resource_type_id = $3, business_system_id = $4,
		name = $5, host = $6, port = $7, credential_id = $8, labels = $9, extra = $10,
		status = $11, updated_at = $12 WHERE instance_id = $1`
	_, err := r.pool.Exec(ctx, query, inst.InstanceID, inst.ZoneID, inst.ResourceTypeID,
		inst.BusinessSysID, inst.Name, inst.Host, inst.Port, inst.CredentialID,
		inst.Labels, inst.Extra, inst.Status, inst.UpdatedAt)
	return err
}

func (r *instanceRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM instance WHERE instance_id = $1`, id)
	return err
}

func (r *instanceRepo) HasTasks(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) > 0 FROM collection_task WHERE instance_id = $1`, id).Scan(&exists)
	return exists, err
}

func (r *instanceRepo) UpdateTestResult(ctx context.Context, id string, testAt time.Time, result string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE instance SET last_test_at = $2, last_test_result = $3, updated_at = $2 WHERE instance_id = $1`,
		id, testAt, result)
	return err
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	s := ""
	for i > 0 {
		s = string(rune('0'+i%10)) + s
		i /= 10
	}
	return s
}
