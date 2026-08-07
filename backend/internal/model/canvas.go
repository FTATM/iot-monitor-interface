package model

import (
	"context"
	"time"
)

type Canvas struct {
	CanvasId   int        `json:"canvasId" db:"canvas_id"`
	CanvasName string     `json:"canvasName" db:"canvas_name"`
	CreatedAt  time.Time  `json:"-" db:"created_at"`
	DeletedAt  *time.Time `json:"-" db:"deleted_at"`
}

type CanvasDetail struct {
	CanvasId   int      `json:"canvasId"`
	CanvasName string   `json:"canvasName"`
	Widgets    []Widget `json:"widgets"`
}

type UpdateCanvas struct {
	CanvasId   int    `json:"canvasId"`
	CanvasName string `json:"canvasName"`
}

type UpsertCanvasRole struct {
	RoleId    int   `json:"roleId"`
	CanvasIds []int `json:"canvasIds"`
}

type CanvasRepository interface {
	GetAll(ctx context.Context, active bool) ([]Canvas, error)
	GetById(ctx context.Context, id int) (*Canvas, error)
	GetByIds(ctx context.Context, ids []int) ([]Canvas, error)
	GetCanvasByUserId(ctx context.Context, userId int, active bool) ([]Canvas, error)
	GetCanvasByRoleId(ctx context.Context, roleId int, active bool) ([]Canvas, error)
	GetAllCanvasRole(ctx context.Context) ([]CanvasRole, error)
	GetCanvasRoleByRoleId(ctx context.Context, roleId int) ([]int, error)
	CreateCanvasRole(ctx context.Context, canvasroles []CanvasRole) error
	DeleteCanvasRole(ctx context.Context, canvasroles []CanvasRole) error
}

type CanvasService interface {
	GetAllCanvas(ctx context.Context) ([]Canvas, error)
	GetCanvasDetailById(ctx context.Context, id int) (*CanvasDetail, error)
	GetAllCanvasDetailByUserRole(ctx context.Context, authUserId int) ([]CanvasDetail, error)
	GetAllCanvasRoleDetail(ctx context.Context) ([]CanvasRoleDetail, error)
	UpsertCanvasRole(ctx context.Context, upsertCanvasRole *UpsertCanvasRole, authUserId int) error
}
