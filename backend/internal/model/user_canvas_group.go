package model

import (
	"context"
)

type UserCanvasGroup struct {
	UserId   int `json:"userId" db:"user_id"`
	CanvasId int `json:"canvasId" db:"canvas_id"`
}

type UserCanvasGroupRepository interface {
	GetByUserId(ctx context.Context, id int) ([]UserCanvasGroup, error)
	GetCanvasByUserId(ctx context.Context, id int) ([]Canvas, error)
	// GetByCanvasId(ctx context.Context, id int) (*Canvas, error)
	// CreateCanvas(ctx context.Context) (*Canvas, error)
}

type UserCanvasGroupService interface {
	GetAllCanvasByUserId(ctx context.Context, id int) ([]UserCanvasGroup, error)
	// GetAllUserByCanvasId(ctx context.Context, id int) (*CanvasDetail, error)
	// CreateCanvas(ctx context.Context) (*Canvas, error)
}
