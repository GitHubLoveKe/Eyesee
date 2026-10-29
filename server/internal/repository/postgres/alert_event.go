package postgres

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type alertEventRepo struct {
	pool *pgxpool.Pool
}

func NewAlertEventRepo(pool *pgxpool.Pool) repository.AlertEventRepo {
	return &alertEventRepo{pool: pool}
}

func (r *alertEventRepo) Create(ctx context.Context, e *model.AlertEvent) error {
	query := `INSERT INTO alert_event (alert_name, zone_id, instance_id, severity, status, labels, annotations, starts_at, ends_at, fingerprint, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING event_id`
	return r.pool.QueryRow(ctx, query, e.AlertName, e.ZoneID, e.InstanceID, e.Severity, e.Status,
		e.Labels, e.Annotations, e.StartsAt, e.EndsAt, e.Fingerprint, time.Now()).Scan(&e.EventID)
}

func (r *alertEventRepo) GetByID(ctx context.Context, id int64) (*model.AlertEvent, error) {
	e := &model.AlertEvent{}
	query := `SELECT event_id, alert_name, zone_id, instance_id, severity, status, labels, annotations, starts_at, ends_at, fingerprint, created_at
		FROM alert_event WHERE event_id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(&e.EventID, &e.AlertName, &e.ZoneID, &e.InstanceID,
		&e.Severity, &e.Status, &e.Labels, &e.Annotations, &e.StartsAt, &e.EndsAt, &e.Fingerprint, &e.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return e, err
}

func (r *alertEventRepo) List(ctx context.Context, f repository.AlertEventFilter) ([]*model.AlertEvent, int64, error) {
	where := " WHERE 1=1"
	var args []interface{}
	argIdx := 1

	if f.ZoneID != "" {
		where += " AND zone_id = $" + itoa(argIdx)
		args = append(args, f.ZoneID)
		argIdx++
	}
	if f.AlertName != "" {
		where += " AND alert_name = $" + itoa(argIdx)
		args = append(args, f.AlertName)
		argIdx++
	}
	if f.Status != nil && len(f.Status) > 0 {
		placeholders := make([]string, len(f.Status))
		for i, s := range f.Status {
			placeholders[i] = "$" + itoa(argIdx)
			args = append(args, s)
			argIdx++
		}
		where += " AND status IN (" + joinPlaceholders(placeholders) + ")"
	}
	if f.Severity != nil && len(f.Severity) > 0 {
		placeholders := make([]string, len(f.Severity))
		for i, s := range f.Severity {
			placeholders[i] = "$" + itoa(argIdx)
			args = append(args, s)
			argIdx++
		}
		where += " AND severity IN (" + joinPlaceholders(placeholders) + ")"
	}
	if f.InstanceID != "" {
		where += " AND instance_id = $" + itoa(argIdx)
		args = append(args, f.InstanceID)
		argIdx++
	}
	if f.Fingerprint != "" {
		where += " AND fingerprint = $" + itoa(argIdx)
		args = append(args, f.Fingerprint)
		argIdx++
	}
	if f.StartTime != nil {
		where += " AND starts_at >= $" + itoa(argIdx)
		args = append(args, *f.StartTime)
		argIdx++
	}
	if f.EndTime != nil {
		where += " AND starts_at <= $" + itoa(argIdx)
		args = append(args, *f.EndTime)
		argIdx++
	}

	var total int64
	countQuery := "SELECT COUNT(*) FROM alert_event" + where
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQuery := `SELECT event_id, alert_name, zone_id, instance_id, severity, status, labels, annotations, starts_at, ends_at, fingerprint, created_at
		FROM alert_event` + where + ` ORDER BY starts_at DESC LIMIT $` + itoa(argIdx) + ` OFFSET $` + itoa(argIdx+1)
	args = append(args, f.Limit, f.Offset)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*model.AlertEvent
	for rows.Next() {
		e := &model.AlertEvent{}
		if err := rows.Scan(&e.EventID, &e.AlertName, &e.ZoneID, &e.InstanceID,
			&e.Severity, &e.Status, &e.Labels, &e.Annotations, &e.StartsAt, &e.EndsAt, &e.Fingerprint, &e.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}

func (r *alertEventRepo) UpdateStatus(ctx context.Context, id int64, status string, endsAt *time.Time) error {
	_, err := r.pool.Exec(ctx, `UPDATE alert_event SET status = $2, ends_at = $3 WHERE event_id = $1`, id, status, endsAt)
	return err
}

func (r *alertEventRepo) Statistics(ctx context.Context, zoneID string, period string) (*repository.AlertStatistics, error) {
	stats := &repository.AlertStatistics{
		BySeverity: make(map[string]int64),
	}

	// Determine time filter
	interval := "24 hours"
	switch period {
	case "1h":
		interval = "1 hour"
	case "6h":
		interval = "6 hours"
	case "7d":
		interval = "7 days"
	}

	timeFilter := " AND starts_at >= NOW() - INTERVAL '" + interval + "'"
	zoneFilter := ""
	var args []interface{}
	argIdx := 1

	if zoneID != "" {
		zoneFilter = " AND zone_id = $" + itoa(argIdx)
		args = append(args, zoneID)
		argIdx++
	}

	// Total firing
	r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM alert_event WHERE status = 'firing'"+timeFilter+zoneFilter, args...).Scan(&stats.TotalFiring)
	// Total resolved
	r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM alert_event WHERE status = 'resolved'"+timeFilter+zoneFilter, args...).Scan(&stats.TotalResolved)

	// By severity
	rows, _ := r.pool.Query(ctx, "SELECT severity, COUNT(*) FROM alert_event WHERE status = 'firing'"+timeFilter+zoneFilter+" GROUP BY severity", args...)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var sev string
			var cnt int64
			rows.Scan(&sev, &cnt)
			stats.BySeverity[sev] = cnt
		}
	}

	// By zone
	rows2, _ := r.pool.Query(ctx, "SELECT zone_id, COUNT(*) FILTER (WHERE status='firing'), COUNT(*) FILTER (WHERE status='resolved') FROM alert_event WHERE 1=1"+timeFilter+" GROUP BY zone_id")
	if rows2 != nil {
		defer rows2.Close()
		for rows2.Next() {
			var z repository.ZoneAlertCount
			rows2.Scan(&z.ZoneID, &z.Firing, &z.Resolved)
			stats.ByZone = append(stats.ByZone, z)
		}
	}

	// Top alerts
	rows3, _ := r.pool.Query(ctx, "SELECT alert_name, COUNT(*) as cnt FROM alert_event WHERE status = 'firing'"+timeFilter+zoneFilter+" GROUP BY alert_name ORDER BY cnt DESC LIMIT 10", args...)
	if rows3 != nil {
		defer rows3.Close()
		for rows3.Next() {
			var a repository.AlertCount
			rows3.Scan(&a.AlertName, &a.Count)
			stats.TopAlerts = append(stats.TopAlerts, a)
		}
	}

	return stats, nil
}

func joinPlaceholders(placeholders []string) string {
	result := ""
	for i, p := range placeholders {
		if i > 0 {
			result += ", "
		}
		result += p
	}
	return result
}
