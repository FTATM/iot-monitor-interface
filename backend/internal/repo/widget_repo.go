package repo

import (
	"context"
	"errors"

	"github.com/FTATM/iot-monitor-interface/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type widgetRepo struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new repository instance
func NewWidgetRepository(db *pgxpool.Pool) model.WidgetRepository {
	return &widgetRepo{db: db}
}

func (r *widgetRepo) GetById(ctx context.Context, id int) (*model.Widget, error) {
	widget := &model.Widget{}
	query := "SELECT widget_id, widget_type_id FROM widget WHERE id = $1"
	err := r.db.QueryRow(ctx, query, id).Scan(&widget.WidgetId, &widget.WidgetTypeId)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("widget not found")
		}
		return nil, err
	}
	return widget, nil
}
