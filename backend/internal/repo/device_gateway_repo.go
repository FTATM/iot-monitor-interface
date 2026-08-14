package repo

import (
	"context"
	"fmt"

	"github.com/FTATM/iot-monitor-interface/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type deviceGatewayRepo struct {
	pool        *pgxpool.Pool
	prefixError string
}

func NewDeviceGatewayRepository(pool *pgxpool.Pool) model.DeviceGatewayRepository {
	return &deviceGatewayRepo{
		pool:        pool,
		prefixError: "deviceGatewayRepo",
	}
}

func (r *deviceGatewayRepo) BulkUpsertDeviceData(ctx context.Context, data []model.DeviceDataRequest) error {
	const fname = "BulkUpsertDeviceData"

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s] begin tx failed: %w", r.prefixError, fname, err)
	}
	defer tx.Rollback(ctx) // no-op if committed

	// 1. Batch update device_latest (device table)
	batch := &pgx.Batch{}
	updateQuery := `UPDATE device SET value_data = $1, updated_at = now(),last_seen_at = now() WHERE device_id = $2`
	for _, d := range data {
		batch.Queue(updateQuery, d.ValueData, d.DeviceId)
	}
	br := tx.SendBatch(ctx, batch)
	for i := range data {
		if _, err := br.Exec(); err != nil {
			br.Close()
			return fmt.Errorf("[%s]>[%s] batch update failed at index %d: %w", r.prefixError, fname, i, err)
		}
	}
	if err := br.Close(); err != nil {
		return fmt.Errorf("[%s]>[%s] batch close failed: %w", r.prefixError, fname, err)
	}

	// 2. Bulk copy insert into device_data_log (history)
	rows := make([][]any, len(data))
	for i, d := range data {
		rows[i] = []any{d.DeviceId, d.ValueData, d.Source, d.ReceivedAt}
	}
	copyCount, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"device_data_log"},
		[]string{"device_id", "value_data", "source", "received_at"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("[%s]>[%s] copy failed: %w", r.prefixError, fname, err)
	}
	if int(copyCount) != len(data) {
		return fmt.Errorf("[%s]>[%s] copy mismatch: expected %d rows, inserted %d", r.prefixError, fname, len(data), copyCount)
	}

	// 3. Commit both together — atomic, either both succeed or both roll back
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s] commit failed: %w", r.prefixError, fname, err)
	}

	return nil
}

func (r *deviceGatewayRepo) Test(ctx context.Context, data []model.DeviceDataRequest) error {
	const fname = "Test"

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s] begin tx failed: %w", r.prefixError, fname, err)
	}
	defer tx.Rollback(ctx) // no-op if committed

	// 1. Batch update device_latest (device table)
	batch := &pgx.Batch{}
	updateQuery := `UPDATE device SET value_data = $1, updated_at = now(),last_seen_at = now() WHERE device_id = $2`
	for _, d := range data {
		batch.Queue(updateQuery, d.ValueData, d.DeviceId)
	}
	br := tx.SendBatch(ctx, batch)
	for i := range data {
		if _, err := br.Exec(); err != nil {
			br.Close()
			return fmt.Errorf("[%s]>[%s] batch update failed at index %d: %w", r.prefixError, fname, i, err)
		}
	}
	if err := br.Close(); err != nil {
		return fmt.Errorf("[%s]>[%s] batch close failed: %w", r.prefixError, fname, err)
	}

	// 3. Commit both together — atomic, either both succeed or both roll back
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s] commit failed: %w", r.prefixError, fname, err)
	}

	return nil
}
