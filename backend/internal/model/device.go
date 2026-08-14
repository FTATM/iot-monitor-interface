package model

import (
	"context"
	"time"
)

type Device struct {
	DeviceId   int        `json:"deviceId" db:"device_id"`
	DeviceName string     `json:"deviceName" db:"device_name"`
	Protocol   string     `json:"protocol" db:"protocol"`
	ValueData  int        `json:"valueData,omitempty" db:"value_data"`
	CreatedAt  time.Time  `json:"-" db:"created_at"`
	DeletedAt  *time.Time `json:"-" db:"deleted_at"`
	IsActive   bool       `json:"isActive" db:"is_active"`
	LastSeenAt *time.Time `json:"lastSeenAt,omitempty" db:"last_seen_at"`
}

func (s *Device) IsSame(req Device) bool {
	if s == nil {
		return false
	}

	return s.DeviceId == req.DeviceId &&
		// s.DeviceName == req.DeviceName &&
		s.IsActive == req.IsActive &&
		s.Protocol == req.Protocol
}

type DeviceDetail struct {
	DeviceId    int        `json:"deviceId"`
	DeviceName  string     `json:"deviceName"`
	Protocol    string     `json:"protocol,omitempty" db:"protocol"`
	ValueData   int        `json:"valueData,omitempty"`
	IsActive    bool       `json:"isActive,omitempty"`
	IsConnected bool       `json:"isConnected,omitempty"`
	LastSeenAt  *time.Time `json:"lastSeenAt,omitempty"`
}

type DeviceCreate struct {
	DeviceName string `json:"deviceName"`
	Protocol   string `json:"protocol"`
	IsActive   bool   `json:"isActive"`
}

type DeviceUpdate struct {
	DeviceId int    `json:"deviceId"`
	Protocol string `json:"protocol"`
	IsActive bool   `json:"isActive"`
}

type ChartDeviceData struct {
	DeviceName string `json:"deviceName"`
	ValueData  int    `json:"valueData"`
}
type ChartData struct {
	DeviceData map[int]ChartDeviceData `json:"deviceData"`
}

type DeviceRepository interface {
	GetById(ctx context.Context, id int) (*Device, error)
	GetAll(ctx context.Context, active bool) ([]Device, error)
	Create(ctx context.Context, device []Device) error
	Update(ctx context.Context, device *Device) error
	Delete(ctx context.Context, deviceId int) error
	GetProtocolType(ctx context.Context) ([]string, error)
	GetByIdChartDeviceData(ctx context.Context, id int) (ChartDeviceData, error)
	GetDeviceDataLogRange(ctx context.Context, deviceIds []int, fromTime, toTime time.Time, maxPoints int) ([]DeviceDataLog, error)

	GetAggregatedData(ctx context.Context, deviceIds []int, fromTime, toTime time.Time, bucketInterval string) ([]DeviceDataLog, error)
	GetRawData(ctx context.Context, deviceIds []int, fromTime, toTime time.Time, limit int) ([]DeviceDataLog, error)
	CountData(ctx context.Context, deviceIds []int, fromTime, toTime time.Time) (int, error)
}

type DeviceService interface {
	GetAllDeviceDetail(ctx context.Context) ([]DeviceDetail, error)
	CreateDevice(ctx context.Context, createDeviceReq []DeviceCreate, authUserId int) error
	UpdateDevice(ctx context.Context, updateDeviceReq *DeviceUpdate, authUserId int) error
	DeleteDevice(ctx context.Context, deviceId, authUserId int) error
	GetProtocolType(ctx context.Context) ([]string, error)
	GetAllDeviceName(ctx context.Context) ([]DeviceDetail, error)
	StartPublic(ctx context.Context)
	AddClient(deviceID string, clientChan chan ChartData)
	RemoveClient(deviceID string, clientChan chan ChartData)
	GetChartHistory(ctx context.Context, deviceId []int, maxPoints int, from, to time.Time) (map[int][][2]float64, error)
}
