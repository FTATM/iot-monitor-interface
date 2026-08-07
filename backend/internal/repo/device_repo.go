package repo

import (
	"context"
	"fmt"

	"github.com/FTATM/iot-monitor-interface/internal/model"
	"github.com/jackc/pgx/v5"
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
	query := "SELECT device_id, device_name, is_active FROM device WHERE device_id = $1"
	err := r.db(ctx).QueryRow(ctx, query, id).Scan(&device.DeviceId, &device.DeviceName, &device.IsActive)

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
		value_data,
		created_at, 
        is_active, 
		is_connected, 
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
			value_data,
			is_active
        ) 
        VALUES ($1, $2, $3) 
        RETURNING device_id
	`

	for _, device := range devices {
		batch.Queue(
			query,
			device.DeviceName,
			device.ValueData,
			device.IsActive,
		)
	}

	br := r.db(ctx).SendBatch(ctx, batch)
	defer br.Close()

	for i := range devices {
		err := br.QueryRow().Scan(&devices[i].DeviceId)
		if err != nil {
			return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
		}
	}

	return nil
}

func (r *deviceRepo) Update(ctx context.Context, device *model.Device) error {
	const fname = "Update"
	query := `
			UPDATE device
			SET 
				is_active = $1
			WHERE device_id = $2
		`
	result, err := r.db(ctx).Exec(ctx, query, device.IsActive, device.DeviceId)

	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
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
