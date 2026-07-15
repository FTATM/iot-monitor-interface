package repo

import (
	"context"

	"github.com/FTATM/iot-monitor-interface/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userCanvasGroupRepo struct {
	pool DBTX
}

func NewUserCanvasGroupRepository(pool *pgxpool.Pool) model.UserCanvasGroupRepository {
	return &userCanvasGroupRepo{pool: pool}
}

func (r *userCanvasGroupRepo) db(ctx context.Context) DBTX {
	if tx := extractTx(ctx); tx != nil {
		return tx // Found a transaction in the context!
	}
	return r.pool // No transaction, use standard pool
}

func (r *userCanvasGroupRepo) GetByUserId(ctx context.Context, id int) ([]model.UserCanvasGroup, error) {
	query := `
		SELECT 
			ucg.user_id, 
			ucg.canvas_id 
		FROM user_canvas_group ucg
		JOIN canvas c on ucg.canvas_id = c.canvas_id
		WHERE user_id = $1
	`
	rows, err := r.db(ctx).Query(ctx, query, id)
	if err != nil {
		return nil, err
	}

	userCanvasGroup, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.UserCanvasGroup])
	if err != nil {
		return nil, err
	}

	return userCanvasGroup, nil
}

func (r *userCanvasGroupRepo) GetCanvasByUserId(ctx context.Context, id int) ([]model.Canvas, error) {
	query := `
		SELECT 
			c.canvas_id,
			c.canvas_name
		FROM user_canvas_group ucg
		JOIN canvas c on ucg.canvas_id = c.canvas_id
		WHERE ucg.user_id = $1
	`
	rows, err := r.db(ctx).Query(ctx, query, id)
	if err != nil {
		return nil, err
	}

	canvasResult, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.Canvas])
	if err != nil {
		return nil, err
	}

	return canvasResult, nil
}
