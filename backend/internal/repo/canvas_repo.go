package repo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/FTATM/iot-monitor-interface/internal/model"
)

type canvasRepo struct {
	db *sql.DB
}

// NewUserRepository creates a new repository instance
func NewCanvasRepository(db *sql.DB) model.CanvasRepository {
	return &canvasRepo{db: db}
}

func (r *canvasRepo) GetById(ctx context.Context, id int) (*model.Canvas, error) {
	// Example DB call
	canvas := &model.Canvas{}
	query := "SELECT canvasId FROM users WHERE id = $1"
	err := r.db.QueryRowContext(ctx, query, id).Scan(&canvas.CanvasId)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("canvas not found")
		}
		return nil, err
	}
	return canvas, nil
}
