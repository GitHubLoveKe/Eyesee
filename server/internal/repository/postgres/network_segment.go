package postgres

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type networkSegmentRepo struct {
	pool *pgxpool.Pool
}

func NewNetworkSegmentRepo(pool *pgxpool.Pool) repository.NetworkSegmentRepo {
	return &networkSegmentRepo{pool: pool}
}

func (r *networkSegmentRepo) Create(ctx context.Context, seg *model.NetworkSegment) error {
	query := `INSERT INTO network_segment (zone_id, cidr, name, description, created_at)
		VALUES ($1, $2, $3, $4, $5) RETURNING segment_id`
	return r.pool.QueryRow(ctx, query, seg.ZoneID, seg.CIDR, seg.Name, seg.Description, seg.CreatedAt).Scan(&seg.SegmentID)
}

func (r *networkSegmentRepo) GetByID(ctx context.Context, id int64) (*model.NetworkSegment, error) {
	seg := &model.NetworkSegment{}
	query := `SELECT segment_id, zone_id, cidr, name, description, created_at
		FROM network_segment WHERE segment_id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&seg.SegmentID, &seg.ZoneID, &seg.CIDR, &seg.Name, &seg.Description, &seg.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return seg, err
}

func (r *networkSegmentRepo) ListByZoneID(ctx context.Context, zoneID string) ([]*model.NetworkSegment, error) {
	query := `SELECT segment_id, zone_id, cidr, name, description, created_at
		FROM network_segment WHERE zone_id = $1 ORDER BY created_at`
	rows, err := r.pool.Query(ctx, query, zoneID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.NetworkSegment
	for rows.Next() {
		s := &model.NetworkSegment{}
		if err := rows.Scan(&s.SegmentID, &s.ZoneID, &s.CIDR, &s.Name, &s.Description, &s.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	return items, nil
}

func (r *networkSegmentRepo) Update(ctx context.Context, seg *model.NetworkSegment) error {
	query := `UPDATE network_segment SET cidr = $2, name = $3, description = $4
		WHERE segment_id = $1`
	_, err := r.pool.Exec(ctx, query, seg.SegmentID, seg.CIDR, seg.Name, seg.Description)
	return err
}

func (r *networkSegmentRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM network_segment WHERE segment_id = $1`, id)
	return err
}

func (r *networkSegmentRepo) CheckCIDROverlap(ctx context.Context, zoneID, cidr string, excludeSegmentID *int64) (bool, error) {
	query := `SELECT COUNT(*) > 0 FROM network_segment WHERE zone_id = $1 AND cidr = $2`
	args := []interface{}{zoneID, cidr}
	if excludeSegmentID != nil {
		query += ` AND segment_id != $3`
		args = append(args, *excludeSegmentID)
	}

	var exists bool
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}
