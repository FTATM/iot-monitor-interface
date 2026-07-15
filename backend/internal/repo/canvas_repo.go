package repo

import (
	"context"
	"errors"

	"github.com/FTATM/iot-monitor-interface/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type canvasRepo struct {
	pool DBTX
}

// NewUserRepository creates a new repository instance
func NewCanvasRepository(pool *pgxpool.Pool) model.CanvasRepository {
	return &canvasRepo{pool: pool}
}

func (r *canvasRepo) db(ctx context.Context) DBTX {
	if tx := extractTx(ctx); tx != nil {
		return tx // Found a transaction in the context!
	}
	return r.pool // No transaction, use standard pool
}

func (r *canvasRepo) GetById(ctx context.Context, id int) (*model.Canvas, error) {
	canvas := &model.Canvas{}
	query := "SELECT canvas_id FROM canvas WHERE canvas_id = $1"
	err := r.db(ctx).QueryRow(ctx, query, id).Scan(&canvas.CanvasId)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("canvas not found")
		}
		return nil, err
	}
	return canvas, nil
}

func (r *canvasRepo) GetByIds(ctx context.Context, ids []int) ([]model.Canvas, error) {
	query := "SELECT canvas_id, canvas_name FROM canvas WHERE canvas_id = ANY($1)"
	rows, err := r.db(ctx).Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}

	canvasList, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.Canvas])
	if err != nil {
		return nil, err
	}

	return canvasList, nil
}
