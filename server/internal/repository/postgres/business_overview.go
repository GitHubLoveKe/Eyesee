package postgres

import (
	"context"
	"dcs/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type businessOverviewRepo struct {
	pool *pgxpool.Pool
}

func NewBusinessOverviewRepo(pool *pgxpool.Pool) repository.BusinessOverviewRepo {
	return &businessOverviewRepo{pool: pool}
}

func (r *businessOverviewRepo) GetOverview(ctx context.Context, systemID string) (*repository.BusinessOverview, error) {
	overview := &repository.BusinessOverview{SystemID: systemID}

	r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM instance WHERE business_system_id = $1`, systemID).Scan(&overview.InstanceCount)
	r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM collection_task ct JOIN instance i ON ct.instance_id = i.instance_id WHERE i.business_system_id = $1`, systemID).Scan(&overview.TaskCount)
	r.pool.QueryRow(ctx, `SELECT COUNT(DISTINCT zone_id) FROM instance WHERE business_system_id = $1`, systemID).Scan(&overview.ZoneCount)

	return overview, nil
}
