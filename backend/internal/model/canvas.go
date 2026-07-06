package model

import "context"

type Canvas struct {
	CanvasId int
}

type CanvasRepository interface {
	// GetById ()
	GetById(ctx context.Context, id int) (*Canvas, error)
}
