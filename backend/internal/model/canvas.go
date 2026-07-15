package model

import (
	"context"
	"time"
)

type Canvas struct {
	CanvasId   int        `json:"canvasId" db:"canvas_id"`
	CanvasName string     `json:"canvasName" db:"canvas_name"`
	CreatedAt  time.Time  `json:"createAt" db:"created_at"`
	UpdatedAt  *time.Time `json:"updateAt" db:"updated_at"`
	DeletedAt  *time.Time `json:"deleteAt" db:"deleted_at"`
}

type CanvasDetail struct {
	CanvasId   int      `json:"canvasId"`
	CanvasName string   `json:"canvasName"`
	Widgets    []Widget `json:"widgets"`
}

type CanvasRepository interface {
	GetById(ctx context.Context, id int) (*Canvas, error)
	GetByIds(ctx context.Context, ids []int) ([]Canvas, error)
	// GetDetailByIds(ctx context.Context, ids []int) ([]CanvasDetail, error)
}

type CanvasService interface {
	GetCanvasDetailById(ctx context.Context, id int) (*CanvasDetail, error)
	GetAllCanvasByUserId(ctx context.Context, id int) ([]CanvasDetail, error)
	// CreateCanvas(ctx context.Context) (*Canvas, error)

}
