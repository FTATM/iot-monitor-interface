package repo

import (
	"context"
	"errors"

	"github.com/FTATM/iot-monitor-interface/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type widgetTypeRepo struct {
	pool DBTX
}

func NewWidgetTypeRepository(pool *pgxpool.Pool) model.WidgetTypeRepository {
	return &widgetTypeRepo{pool: pool}
}

func (r *widgetTypeRepo) db(ctx context.Context) DBTX {
	if tx := extractTx(ctx); tx != nil {
		return tx // Found a transaction in the context!
	}
	return r.pool // No transaction, use standard pool
}

func (r *widgetTypeRepo) GetById(ctx context.Context, id int) (*model.WidgetType, error) {
	widget := &model.WidgetType{}
	query := "SELECT widget_type_id, widget_type_name FROM widget_type WHERE widget_type_id = $1"
	err := r.db(ctx).QueryRow(ctx, query, id).Scan(&widget.WidgetTypeId, &widget.WidgetTypeName)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("widget not found")
		}
		return nil, err
	}
	return widget, nil
}

func (r *widgetTypeRepo) GetAll(ctx context.Context) ([]model.WidgetType, error) {
	query := "SELECT widget_type_id, widget_type_name FROM widget_type"
	rows, err := r.db(ctx).Query(ctx, query)
	if err != nil {
		return nil, err
	}

	widgetTypes, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.WidgetType])
	if err != nil {
		return nil, err
	}

	return widgetTypes, nil
}
