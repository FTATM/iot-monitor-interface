package model

import "context"

type Widget struct {
	WidgetId     string `json:"i"`
	WidgetTypeId string `json:"widgetTypeId"`
	X            int    `json:"x"`
	Y            int    `json:"y"`
	W            int    `json:"w"`
	H            int    `json:"h"`
}

type WidgetRepository interface {
	GetById(ctx context.Context, id int) (*Widget, error)
}
