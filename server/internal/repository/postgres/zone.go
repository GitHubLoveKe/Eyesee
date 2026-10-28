package postgres

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type zoneRepo struct {
	pool *pgxpool.Pool
}

func NewZoneRepo(pool *pgxpool.Pool) repository.ZoneRepo {
	return &zoneRepo{pool: pool}
}

func (r *zoneRepo) Create(ctx context.Context, z *model.Zone) error {
	query := `INSERT INTO zone (zone_id, name, description, collector_type, remote_write_url, default_sink_id, config, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.pool.Exec(ctx, query, z.ZoneID, z.Name, z.Description, z.CollectorType, z.RemoteWriteURL, z.DefaultSinkID, z.Config, z.Status, z.CreatedAt, z.UpdatedAt)
	return err
}

func (r *zoneRepo) GetByID(ctx context.Context, id string) (*model.Zone, error) {
	z := &model.Zone{}
	query := `SELECT zone_id, name, description, collector_type, remote_write_url, default_sink_id, config, status, created_at, updated_at
		FROM zone WHERE zone_id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&z.ZoneID, &z.Name, &z.Description, &z.CollectorType, &z.RemoteWriteURL,
		&z.DefaultSinkID, &z.Config, &z.Status, &z.CreatedAt, &z.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return z, err
}

func (r *zoneRepo) List(ctx context.Context, status string, offset, limit int) ([]*model.Zone, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM zone`
	listQuery := `SELECT zone_id, name, description, collector_type, remote_write_url, default_sink_id, config, status, created_at, updated_at
		FROM zone ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	var args []interface{}
	if status != "" {
		countQuery += ` WHERE status = $1`
		listQuery = `SELECT zone_id, name, description, collector_type, remote_write_url, default_sink_id, config, status, created_at, updated_at
			FROM zone WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		args = append(args, status)
	}

	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*model.Zone
	for rows.Next() {
		z := &model.Zone{}
		if err := rows.Scan(&z.ZoneID, &z.Name, &z.Description, &z.CollectorType, &z.RemoteWriteURL,
			&z.DefaultSinkID, &z.Config, &z.Status, &z.CreatedAt, &z.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, z)
	}
	return items, total, nil
}

func (r *zoneRepo) Update(ctx context.Context, z *model.Zone) error {
	query := `UPDATE zone SET name = $2, description = $3, collector_type = $4, remote_write_url = $5,
		default_sink_id = $6, config = $7, status = $8, updated_at = $9 WHERE zone_id = $1`
	_, err := r.pool.Exec(ctx, query, z.ZoneID, z.Name, z.Description, z.CollectorType,
		z.RemoteWriteURL, z.DefaultSinkID, z.Config, z.Status, z.UpdatedAt)
	return err
}

func (r *zoneRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM zone WHERE zone_id = $1`, id)
	return err
}

func (r *zoneRepo) CreateSegment(ctx context.Context, s *model.NetworkSegment) error {
	query := `INSERT INTO network_segment (zone_id, cidr, name, description, created_at)
		VALUES ($1, $2, $3, $4, $5) RETURNING segment_id`
	return r.pool.QueryRow(ctx, query, s.ZoneID, s.CIDR, s.Name, s.Description, s.CreatedAt).Scan(&s.SegmentID)
}

func (r *zoneRepo) ListSegments(ctx context.Context, zoneID string) ([]*model.NetworkSegment, error) {
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

func (r *zoneRepo) UpdateSegment(ctx context.Context, s *model.NetworkSegment) error {
	query := `UPDATE network_segment SET cidr = $2, name = $3, description = $4
		WHERE segment_id = $1`
	_, err := r.pool.Exec(ctx, query, s.SegmentID, s.CIDR, s.Name, s.Description)
	return err
}

func (r *zoneRepo) DeleteSegment(ctx context.Context, segmentID int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM network_segment WHERE segment_id = $1`, segmentID)
	return err
}

func (r *zoneRepo) CheckCIDROverlap(ctx context.Context, zoneID, cidr string, excludeSegmentID *int64) (bool, error) {
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

func parseCIDRHost(cidr string) string {
	if idx := strings.Index(cidr, "/"); idx >= 0 {
		return cidr[:idx]
	}
	return cidr
}
