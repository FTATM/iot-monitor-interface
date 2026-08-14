package model

import (
	"context"
)

type Widget struct {
	WidgetId        int         `json:"widgetId" db:"widget_id"`
	WidgetTypeId    int         `json:"widgetTypeId" db:"widget_type_id"`
	CanvasId        int         `json:"canvasId" db:"canvas_id"`
	DeviceIds       []int       `json:"deviceIds" db:"device_id_s"`
	WidgetLabel     string      `json:"widgetLabel" db:"widget_label"`
	LayoutData      DynamicJSON `json:"layoutData" db:"layout_data"`
	WidgetColor     DynamicJSON `json:"widgetColor" db:"widget_color"`
	CustomChartData DynamicJSON `json:"customChartData" db:"custom_chart_data"`
}

type UpsertWidgetReqest struct {
	CanvasId      int            `json:"canvasId"`
	UpsertWidgets []UpsertWidget `json:"upsertWidgets"`
}

type UpsertWidget struct {
	WidgetId        int         `json:"widgetId"`
	WidgetTypeId    int         `json:"widgetTypeId"`
	DeviceIds       []int       `json:"deviceIds"`
	WidgetLabel     string      `json:"widgetLabel"`
	LayoutData      DynamicJSON `json:"layoutData"`
	WidgetColor     DynamicJSON `json:"widgetColor"`
	CustomChartData DynamicJSON `json:"customChartData"`
}

type WidgetRepository interface {
	GetById(ctx context.Context, id int) (*Widget, error)
	Create(ctx context.Context, widgets []Widget) error
	Update(ctx context.Context, widgets []Widget) error
	Delete(ctx context.Context, ids []int) error
	GetWidgetByCanvasId(ctx context.Context, canvasId []int) ([]Widget, error)
}

type WidgetService interface {
	GetWidgetDetailById(ctx context.Context, widgetId int) (*Widget, error)
	CreateWidget(ctx context.Context, widgets []Widget) error
	UpdateWidget(ctx context.Context, widgets []Widget) error
	DeleteWidget(ctx context.Context, widgetIds []int) error
	UpsertWidget(ctx context.Context, req *UpsertWidgetReqest) error
}
