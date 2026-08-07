package model

import (
	"context"
	"time"
)

type Device struct {
	DeviceId    int        `json:"deviceId" db:"device_id"`
	DeviceName  string     `json:"deviceName" db:"device_name"`
	ValueData   int        `json:"valueData,omitempty" db:"value_data"`
	CreatedAt   time.Time  `json:"-" db:"created_at"`
	DeletedAt   *time.Time `json:"-" db:"deleted_at"`
	IsActive    bool       `json:"isActive" db:"is_active"`
	IsConnected bool       `json:"isConnected,omitempty" db:"is_connected"`
	LastSeenAt  *time.Time `json:"lastSeenAt,omitempty" db:"last_seen_at"`
}

func (s *Device) IsSame(req Device) bool {
	if s == nil {
		return false
	}

	return s.DeviceId == req.DeviceId &&
		// s.DeviceName == req.DeviceName &&
		s.IsActive == req.IsActive
}

type DeviceDetail struct {
	DeviceId    int        `json:"deviceId"`
	DeviceName  string     `json:"deviceName"`
	ValueData   int        `json:"valueData"`
	IsActive    bool       `json:"isActive"`
	IsConnected bool       `json:"isConnected"`
	LastSeenAt  *time.Time `json:"lastSeenAt"`
}

type DeviceCreate struct {
	DeviceName string `json:"deviceName"`
	IsActive   bool   `json:"isActive"`
}

type DeviceUpdate struct {
	DeviceId int  `json:"deviceId"`
	IsActive bool `json:"isActive"`
}

type DeviceRepository interface {
	GetById(ctx context.Context, id int) (*Device, error)
	GetAll(ctx context.Context, active bool) ([]Device, error)
	Create(ctx context.Context, device []Device) error
	Update(ctx context.Context, device *Device) error
	Delete(ctx context.Context, deviceId int) error
}

type DeviceService interface {
	GetAllDeviceDetail(ctx context.Context) ([]DeviceDetail, error)
	CreateDevice(ctx context.Context, createDeviceReq []DeviceCreate, authUserId int) error
	UpdateDevice(ctx context.Context, updateDeviceReq *DeviceUpdate, authUserId int) error
	DeleteDevice(ctx context.Context, deviceId, authUserId int) error
}
