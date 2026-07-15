package model

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type Widget struct {
	WidgetId     int        `json:"widgetId" db:"widget_id"`
	WidgetTypeId int        `json:"widgetTypeId" db:"widget_type_id"`
	CanvasId     int        `json:"canvasId" db:"canvas_id"`
	LayoutData   LayoutData `json:"layoutData" db:"layout_data"`
}

type LayoutData struct {
	X int `json:"x" db:"x"`
	Y int `json:"y" db:"y"`
	W int `json:"w" db:"w"`
	H int `json:"h" db:"h"`
}

type WidgetRepository interface {
	GetById(ctx context.Context, id int) (*Widget, error)
	Create(ctx context.Context, widgets []Widget) error
	Update(ctx context.Context, widgets []Widget) error
	Delete(ctx context.Context, ids []int) error
	GetWidgetByCanvasId(ctx context.Context, canvasId int) ([]Widget, error)
	GetWidgetByCanvasIds(ctx context.Context, canvasId []int) ([]Widget, error)
}

type WidgetService interface {
	GetWidgetDetailById(ctx context.Context, widgetID int) (*Widget, error)
	CreateWidgets(ctx context.Context, widgets []Widget) error
	UpdateWidget(ctx context.Context, widgets []Widget) error
	DeleteWidgets(ctx context.Context, ids []int) error
}

func (l *LayoutData) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	// pgx usually passes JSONB as []byte or string
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, l)
	case string:
		return json.Unmarshal([]byte(v), l)
	default:
		return errors.New("type assertion to []byte/string failed for LayoutData")
	}
}

// Optional: Make it implement driver.Valuer for INSERTS/UPDATES
func (l LayoutData) Value() (driver.Value, error) {
	return json.Marshal(l)
}
