package repo

import (
	"context"
	"fmt"

	"github.com/FTATM/iot-monitor-interface/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type logReportRepo struct {
	pool        DBTX
	prefixError string
}

func NewLogReportRepository(pool *pgxpool.Pool) model.LogReportRepository {
	return &logReportRepo{pool: pool, prefixError: "logReportRepo"}
}

func (r *logReportRepo) db(ctx context.Context) DBTX {
	if tx := extractTx(ctx); tx != nil {
		return tx
	}
	return r.pool
}

func (r *logReportRepo) GetSystemLogsForExport(ctx context.Context, filter model.LogFilter) ([]model.AuditLogReport, error) {
	const fname = "GetSystemLogsForExport"
	query := `
		SELECT 
			a.id, a.entity_type, a.entity_id, a.action, a.changed_by, 
			COALESCE(u.username, 'System') AS username, 
			a.old_data, a.new_data, a.created_at
		FROM audit_log a
		LEFT JOIN "user" u ON a.changed_by = u.user_id
		WHERE a.created_at >= $1 AND a.created_at <= $2
			AND (cardinality($3::text[]) = 0 OR a.entity_type = ANY($3::TEXT[]))
			-- ⚡ NEW: Search across Entity ID and Username
			AND ($4 = '' OR a.entity_id ILIKE '%' || $4 || '%' OR u.username ILIKE '%' || $4 || '%')
		ORDER BY a.created_at DESC
	`
	// ⚡ Include filter.Keyword as the 4th argument
	rows, err := r.db(ctx).Query(ctx, query, filter.From, filter.To, filter.EntityTypes, filter.Keyword)

	logs, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.AuditLogReport])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return logs, nil
}

func (r *logReportRepo) GetDeviceLogsForExport(ctx context.Context, filter model.LogFilter) ([]model.DeviceDataLogReport, error) {
	const fname = "GetDeviceLogsForExport"

	query := `
		SELECT 
			l.device_id, 
			(l.value_data / $4::numeric) AS value_data,
			l.received_at,
			COALESCE(d.device_name, 'Unknown Device') AS device_name
		FROM device_data_log l
		LEFT JOIN device d ON l.device_id = d.device_id
		WHERE l.received_at >= $1 AND l.received_at <= $2
		  AND (
			$3 = '' OR 
			l.device_id::text ILIKE '%' || $3 || '%' OR
			d.device_name ILIKE '%' || $3 || '%' OR
			(l.value_data / $4::numeric)::text ILIKE '%' || $3 || '%'
		  )
		ORDER BY l.received_at DESC
	`

	// ⚡ FIX: Pass the constant as the 4th parameter
	rows, err := r.db(ctx).Query(ctx, query, filter.From, filter.To, filter.Keyword, float64(model.DeviceScale))
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	logs, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.DeviceDataLogReport])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return logs, nil
}

func (r *logReportRepo) SearchSystemLogs(ctx context.Context, filter model.LogFilter) ([]model.AuditLogReport, error) {
	const fname = "SearchSystemLogs"

	// ⚡ 1. Build a strict, SQL-Injection-proof Order By string
	orderBy := "a.created_at DESC" // Default fallback
	if filter.SortBy != "" {
		dir := "ASC"
		if filter.SortDesc {
			dir = "DESC"
		}

		// Map the Vue accessorKey names to real database columns
		switch filter.SortBy {
		case "createdAt":
			orderBy = "a.created_at " + dir
		case "action":
			orderBy = "a.action " + dir
		case "entityType":
			orderBy = "a.entity_type " + dir
		case "entityId":
			orderBy = fmt.Sprintf("LENGTH(a.entity_id) %s, a.entity_id %s", dir, dir)
		case "username":
			orderBy = "u.username " + dir
		}
	}

	query := fmt.Sprintf(`
		SELECT 
			a.id, a.entity_type, a.entity_id, a.action, a.changed_by, 
			COALESCE(u.username, 'System') AS username, 
			a.old_data, a.new_data, a.created_at
		FROM audit_log a
		LEFT JOIN "user" u ON a.changed_by = u.user_id
		WHERE a.created_at >= $1 AND a.created_at <= $2
		  AND (cardinality($3::text[]) = 0 OR a.entity_type = ANY($3::TEXT[]))
		  AND ($4 = '' OR a.entity_id ILIKE '%%' || $4 || '%%' OR u.username ILIKE '%%' || $4 || '%%')
		ORDER BY %s 
		LIMIT $5 OFFSET $6
	`, orderBy)

	rows, err := r.db(ctx).Query(ctx, query, filter.From, filter.To, filter.EntityTypes, filter.Keyword, filter.Limit, filter.Offset)

	logs, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.AuditLogReport])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return logs, nil
}

func (r *logReportRepo) SearchDeviceLogs(ctx context.Context, filter model.LogFilter) ([]model.DeviceDataLogReport, error) {
	const fname = "SearchDeviceLogs"

	query := `
		SELECT 
			l.device_id,
			(l.value_data / $6::numeric) AS value_data,
			l.received_at,
			COALESCE(d.device_name, 'Unknown Device') AS device_name
		FROM device_data_log l
		LEFT JOIN device d ON l.device_id = d.device_id
		WHERE l.received_at >= $1 AND l.received_at <= $2
		  AND (
			$3 = '' OR 
			l.device_id::text ILIKE '%' || $3 || '%' OR
			d.device_name ILIKE '%' || $3 || '%' OR
			(l.value_data / $6::numeric)::text ILIKE '%' || $3 || '%'
		  )
		ORDER BY l.received_at DESC
		LIMIT $4 OFFSET $5
	`

	rows, err := r.db(ctx).Query(ctx, query, filter.From, filter.To, filter.Keyword, filter.Limit, filter.Offset, float64(model.DeviceScale))
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	logs, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.DeviceDataLogReport])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return logs, nil
}

func (r *logReportRepo) GetAuditLogEntityTypes(ctx context.Context) ([]string, error) {
	const fname = "GetAuditLogEntityTypes"
	query := `SELECT entity_type FROM audit_log GROUP BY entity_type ORDER BY entity_type`

	rows, err := r.db(ctx).Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	types, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return types, nil
}

func (r *logReportRepo) CountSystemLogs(ctx context.Context, filter model.LogFilter) (int, error) {
	const fname = "CountSystemLogs"
	query := `
		SELECT COUNT(*)
		FROM audit_log a
		-- ⚡ NEW: Must add the join so we can count by username
		LEFT JOIN "user" u ON a.changed_by = u.user_id
		WHERE a.created_at >= $1 AND a.created_at <= $2
		  AND (cardinality($3::text[]) = 0 OR a.entity_type = ANY($3::TEXT[]))
		  -- ⚡ NEW: Search across Entity ID and Username
		  AND ($4 = '' OR a.entity_id ILIKE '%' || $4 || '%' OR u.username ILIKE '%' || $4 || '%')
	`
	var count int
	// ⚡ Include filter.Keyword as the 4th argument
	err := r.db(ctx).QueryRow(ctx, query, filter.From, filter.To, filter.EntityTypes, filter.Keyword).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return count, nil
}

func (r *logReportRepo) CountDeviceLogs(ctx context.Context, filter model.LogFilter) (int, error) {
	const fname = "CountDeviceLogs"
	query := `
		SELECT COUNT(*)
		FROM device_data_log l
		LEFT JOIN device d ON l.device_id = d.device_id
		WHERE l.received_at >= $1 AND l.received_at <= $2
		  AND (
			$3 = '' OR 
			l.device_id::text ILIKE '%' || $3 || '%' OR
			d.device_name ILIKE '%' || $3 || '%' OR
			(l.value_data / $4::numeric)::text ILIKE '%' || $3 || '%'
		  )
	`
	var count int
	// ⚡ FIX: Pass the constant as the 4th parameter
	err := r.db(ctx).QueryRow(ctx, query, filter.From, filter.To, filter.Keyword, float64(model.DeviceScale)).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return count, nil
}
