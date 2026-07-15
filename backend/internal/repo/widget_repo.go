package repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/FTATM/iot-monitor-interface/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type widgetRepo struct {
	pool DBTX
}

// NewUserRepository creates a new repository instance
func NewWidgetRepository(pool *pgxpool.Pool) model.WidgetRepository {
	return &widgetRepo{pool: pool}
}

func (r *widgetRepo) db(ctx context.Context) DBTX {
	if tx := extractTx(ctx); tx != nil {
		return tx // Found a transaction in the context!
	}
	return r.pool // No transaction, use standard pool
}

func (r *widgetRepo) GetById(ctx context.Context, id int) (*model.Widget, error) {
	widget := &model.Widget{}
	query := "SELECT widget_id, widget_type_id , layout_data FROM widget WHERE widget_id = $1"
	err := r.db(ctx).QueryRow(ctx, query, id).Scan(&widget.WidgetId, &widget.WidgetTypeId, &widget.LayoutData)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("widget not found")
		}
		return nil, err
	}
	return widget, nil
}

func (r *widgetRepo) Create(ctx context.Context, widgets []model.Widget) error {
	if len(widgets) == 0 {
		return nil
	}

	var values []any
	var placeholders []string

	for i, widget := range widgets {

		rowValues := []any{
			widget.WidgetTypeId,
			widget.CanvasId,
			widget.LayoutData,
		}

		numCols := len(rowValues)
		var rowPlaceholders []string

		for j := 0; j < numCols; j++ {
			paramIndex := (i * numCols) + j + 1
			rowPlaceholders = append(rowPlaceholders, fmt.Sprintf("$%d", paramIndex))
		}

		rowString := fmt.Sprintf("(%s)", strings.Join(rowPlaceholders, ", "))
		placeholders = append(placeholders, rowString)

		values = append(values, rowValues...)
	}

	query := fmt.Sprintf(
		`INSERT INTO widget (widget_type_id, canvas_id, layout_data) VALUES %s RETURNING widget_id`,
		strings.Join(placeholders, ", "),
	)
	rows, err := r.db(ctx).Query(ctx, query, values...)
	if err != nil {
		return err
	}
	defer rows.Close()

	i := 0
	for rows.Next() {
		if err := rows.Scan(&widgets[i].WidgetId); err != nil {
			return err
		}
		i++
	}

	return rows.Err()
}

func (r *widgetRepo) Update(ctx context.Context, widgets []model.Widget) error {
	if len(widgets) == 0 {
		return nil
	}

	// 1. Create slices to hold the columns of data we want to update
	var ids []int
	var typeIDs []int
	var layouts []string // Using string is the safest way to pass JSONB arrays to Postgres

	// 2. Populate the slices from our structs
	for _, w := range widgets {
		ids = append(ids, w.WidgetId)
		typeIDs = append(typeIDs, w.WidgetTypeId)

		layoutBytes, err := json.Marshal(w.LayoutData)
		if err != nil {
			return err
		}
		layouts = append(layouts, string(layoutBytes))
	}

	query := `UPDATE widget AS w
		SET 
			widget_type_id = u.widget_type_id,
			layout_data = u.layout_data,
			updated_at = CURRENT_TIMESTAMP
		FROM UNNEST($1::int[], $2::int[], $3::jsonb[]) AS u(widget_id, widget_type_id, layout_data)
		WHERE w.widget_id = u.widget_id`

	result, err := r.db(ctx).Exec(ctx, query, ids, typeIDs, layouts)

	if err != nil {
		return err
	}

	if result.RowsAffected() != int64(len(ids)) {
		return fmt.Errorf("update row affected not match")
	}

	return nil
}

func (r *widgetRepo) Delete(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}

	query := `DELETE FROM widget WHERE widget_id = ANY($1::int[])`

	result, err := r.db(ctx).Exec(ctx, query, ids)
	if err != nil {
		return err
	}

	if result.RowsAffected() != int64(len(ids)) {
		return fmt.Errorf("attempted to delete %d widgets, but only found and deleted %d", len(ids), result.RowsAffected())
	}

	return nil
}

func (r *widgetRepo) GetWidgetByCanvasId(ctx context.Context, canvasId int) ([]model.Widget, error) {
	query := "SELECT widget_id, widget_type_id, canvas_id, layout_data FROM widget WHERE canvas_id = $1"
	rows, err := r.db(ctx).Query(ctx, query, canvasId)
	if err != nil {
		return nil, err
	}

	widgets, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.Widget])
	if err != nil {
		return nil, err
	}

	return widgets, nil
}

func (r *widgetRepo) GetWidgetByCanvasIds(ctx context.Context, canvasId []int) ([]model.Widget, error) {
	query := "SELECT widget_id, widget_type_id, canvas_id, layout_data FROM widget WHERE canvas_id = ANY($1)"
	rows, err := r.db(ctx).Query(ctx, query, canvasId)
	if err != nil {
		return nil, err
	}

	widgets, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.Widget])
	if err != nil {
		return nil, err
	}

	return widgets, nil
}
