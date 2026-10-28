package postgres

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type resourceTypeRepo struct {
	pool *pgxpool.Pool
}

func NewResourceTypeRepo(pool *pgxpool.Pool) repository.ResourceTypeRepo {
	return &resourceTypeRepo{pool: pool}
}

func (r *resourceTypeRepo) Create(ctx context.Context, rt *model.ResourceType) error {
	query := `INSERT INTO resource_type (type_id, parent_id, name, description, icon, sort_order, attributes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.pool.Exec(ctx, query, rt.TypeID, rt.ParentID, rt.Name, rt.Description, rt.Icon, rt.SortOrder, rt.Attributes, rt.CreatedAt, rt.UpdatedAt)
	return err
}

func (r *resourceTypeRepo) GetByID(ctx context.Context, id string) (*model.ResourceType, error) {
	rt := &model.ResourceType{}
	query := `SELECT type_id, parent_id, name, description, icon, sort_order, attributes, created_at, updated_at
		FROM resource_type WHERE type_id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&rt.TypeID, &rt.ParentID, &rt.Name, &rt.Description, &rt.Icon, &rt.SortOrder,
		&rt.Attributes, &rt.CreatedAt, &rt.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return rt, err
}

func (r *resourceTypeRepo) List(ctx context.Context, parentID *string) ([]*model.ResourceType, error) {
	var rows pgx.Rows
	var err error

	if parentID != nil {
		query := `SELECT type_id, parent_id, name, description, icon, sort_order, attributes, created_at, updated_at
			FROM resource_type WHERE parent_id = $1 ORDER BY sort_order, name`
		rows, err = r.pool.Query(ctx, query, *parentID)
	} else {
		query := `SELECT type_id, parent_id, name, description, icon, sort_order, attributes, created_at, updated_at
			FROM resource_type ORDER BY sort_order, name`
		rows, err = r.pool.Query(ctx, query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.ResourceType
	for rows.Next() {
		rt := &model.ResourceType{}
		if err := rows.Scan(&rt.TypeID, &rt.ParentID, &rt.Name, &rt.Description, &rt.Icon,
			&rt.SortOrder, &rt.Attributes, &rt.CreatedAt, &rt.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, rt)
	}
	return items, nil
}

func (r *resourceTypeRepo) Update(ctx context.Context, rt *model.ResourceType) error {
	query := `UPDATE resource_type SET parent_id = $2, name = $3, description = $4, icon = $5,
		sort_order = $6, attributes = $7, updated_at = $8 WHERE type_id = $1`
	_, err := r.pool.Exec(ctx, query, rt.TypeID, rt.ParentID, rt.Name, rt.Description,
		rt.Icon, rt.SortOrder, rt.Attributes, rt.UpdatedAt)
	return err
}

func (r *resourceTypeRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM resource_type WHERE type_id = $1`, id)
	return err
}

func (r *resourceTypeRepo) HasChildren(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) > 0 FROM resource_type WHERE parent_id = $1`, id).Scan(&exists)
	return exists, err
}

func (r *resourceTypeRepo) HasInstances(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) > 0 FROM instance WHERE resource_type_id = $1`, id).Scan(&exists)
	return exists, err
}
