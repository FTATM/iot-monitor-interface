package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/FTATM/iot-monitor-interface/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type deviceRepo struct {
	pool        DBTX
	prefixError string
}

func NewDeviceRepository(pool *pgxpool.Pool) model.DeviceRepository {
	return &deviceRepo{pool: pool, prefixError: "deviceRepo"}
}

func (r *deviceRepo) db(ctx context.Context) DBTX {
	if tx := extractTx(ctx); tx != nil {
		return tx
	}
	return r.pool
}

func (r *deviceRepo) GetById(ctx context.Context, id int) (*model.Device, error) {
	const fname = "GetById"
	device := &model.Device{}
	query := "SELECT device_id, device_name, is_active, protocol FROM device WHERE device_id = $1"
	err := r.db(ctx).QueryRow(ctx, query, id).Scan(&device.DeviceId, &device.DeviceName, &device.IsActive, &device.Protocol)

	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return device, nil
}

func (r *deviceRepo) GetAll(ctx context.Context, active bool) ([]model.Device, error) {
	const fname = "GetAll"
	query := `
	SELECT 
		device_id, 
		device_name, 
		protocol,
		value_data,
		created_at, 
        is_active, 
		last_seen_at 
    FROM device
	WHERE ($1 = false) OR ($1 = true AND deleted_at IS NULL)
	`
	rows, err := r.db(ctx).Query(ctx, query, active)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	devices, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.Device])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return devices, nil
}

func (r *deviceRepo) Create(ctx context.Context, devices []model.Device) error {
	const fname = "Create"
	if len(devices) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `
        INSERT INTO device (
			device_name,
			protocol,
			is_active
        ) 
        VALUES ($1, $2,$3) 
        RETURNING device_id
	`

	for _, device := range devices {
		batch.Queue(
			query,
			device.DeviceName,
			device.Protocol,
			device.IsActive,
		)
	}

	br := r.db(ctx).SendBatch(ctx, batch)
	defer br.Close()

	for i := range devices {
		err := br.QueryRow().Scan(&devices[i].DeviceId)
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" {
				return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, model.ErrDuplicate)
			} else {
				return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
			}
		}
	}

	return nil
}

func (r *deviceRepo) Update(ctx context.Context, device *model.Device) error {
	const fname = "Update"
	query := `
			UPDATE device
			SET 
				is_active = $1,
				protocol = $3
			WHERE device_id = $2
		`
	result, err := r.db(ctx).Exec(ctx, query, device.IsActive, device.DeviceId, device.Protocol)

	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		if pgErr.Code == "23505" {
			return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, model.ErrDuplicate)
		} else {
			return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
		}
	}

	if result.RowsAffected() != 1 {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, pgx.ErrNoRows)
	}

	return nil

}

func (r *deviceRepo) Delete(ctx context.Context, deviceId int) error {
	const fname = "Delete"
	query := `
			UPDATE device
			SET 
				is_active = false,
				deleted_at = CURRENT_TIMESTAMP
			WHERE device_id = $1
		`
	result, err := r.db(ctx).Exec(ctx, query, deviceId)

	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	if result.RowsAffected() != 1 {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, pgx.ErrNoRows)
	}

	return nil

}

func (r *deviceRepo) GetProtocolType(ctx context.Context) ([]string, error) {
	const fname = "GetAll"
	query := `
		SELECT unnest(enum_range(NULL::device_protocol_type)) AS protocol_name;
	`
	rows, err := r.db(ctx).Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	protocols, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return protocols, nil
}

func (r *deviceRepo) GetByIdChartDeviceData(ctx context.Context, id int) (model.ChartDeviceData, error) {
	const fname = "GetByIdChartDevice"
	device := model.ChartDeviceData{}
	query := "SELECT device_name, value_data FROM device WHERE device_id = $1"
	err := r.db(ctx).QueryRow(ctx, query, id).Scan(&device.DeviceName, &device.ValueData)

	if err != nil {
		return device, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return device, nil
}

func (r *deviceRepo) GetDeviceDataLogRange(ctx context.Context, deviceIds []int, fromTime, toTime time.Time, maxPoints int) ([]model.DeviceDataLog, error) {
	const fname = "GetDeviceDataLogRange"
	query := `
		SELECT 
			device_id, 
			received_at, 
			value_data
		FROM 
			device_data_log
		WHERE 
			device_id = ANY($1::int[]) 
			AND received_at >= $2 
			AND received_at <= $3
		ORDER BY 
			received_at ASC
		LIMIT $4;
	`

	rows, err := r.db(ctx).Query(ctx, query, deviceIds, fromTime, toTime, maxPoints)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	logs, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.DeviceDataLog])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return logs, nil
}

func (r *deviceRepo) CountData(ctx context.Context, deviceIds []int, fromTime, toTime time.Time) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM device_data_log 
		WHERE device_id = ANY($1::int[]) 
		  AND received_at >= $2 
		  AND received_at <= $3;
	`
	var count int
	err := r.db(ctx).QueryRow(ctx, query, deviceIds, fromTime, toTime).Scan(&count)
	return count, err
}

// 2. Fetch Pure Raw Data (No bucketing, exactly as inserted)
func (r *deviceRepo) GetRawData(ctx context.Context, deviceIds []int, fromTime, toTime time.Time, limit int) ([]model.DeviceDataLog, error) {
	query := `
		SELECT device_id, received_at, value_data
		FROM device_data_log
		WHERE device_id = ANY($1::int[]) AND received_at >= $2 AND received_at <= $3
		ORDER BY received_at ASC
		LIMIT $4;
	`
	rows, err := r.db(ctx).Query(ctx, query, deviceIds, fromTime, toTime, limit)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.DeviceDataLog])
}

// 3. Fetch Aggregated Data (Using time_bucket and last() to protect the browser)
func (r *deviceRepo) GetAggregatedData(ctx context.Context, deviceIds []int, fromTime, toTime time.Time, bucketInterval string) ([]model.DeviceDataLog, error) {
	query := `
		SELECT 
			device_id, 
			time_bucket($4::interval, received_at) AS received_at, 
			last(value_data, received_at)::INT AS value_data
		FROM device_data_log
		WHERE device_id = ANY($1::int[]) AND received_at >= $2 AND received_at <= $3
		GROUP BY device_id, time_bucket($4::interval, received_at)
		ORDER BY received_at ASC;
	`
	rows, err := r.db(ctx).Query(ctx, query, deviceIds, fromTime, toTime, bucketInterval)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.DeviceDataLog])
}
