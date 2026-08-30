package model

import (
	"context"
	"time"
)

type DeviceDataPayloadReq struct {
	DeviceName int `json:"deviceName,omitempty"`
	ValueData  int `json:"valueData"`
}

type DeviceData struct {
	DeviceId   int       `json:"deviceId" db:"device_id"`
	ValueData  int       `json:"valueData" db:"value_data"`
	ReceivedAt time.Time `json:"-"`
}

type GatewayCommand struct {
	DeviceId int             `json:"deviceId"`
	GroupId  int             `json:"groupId"`
	Payload  []DeviceCommand `json:"payload"`
	Protocol string          `json:"protocol"`
}

type DeviceCommand struct {
	DeviceName string `json:"deviceName"`
	Cmd        string `json:"cmd"`
}

type DeviceGatewayRepository interface {
	BulkUpsertDeviceData(ctx context.Context, data []DeviceData) error
	UpdateLastSeen(ctx context.Context, deviceId int) error
	GetDeviceIdByName(ctx context.Context, deviceName string) (int, error)
	GetDeviceIdByGroupName(ctx context.Context, deviceName string) ([]DeviceGroupData, error)
	GetDeviceNameById(ctx context.Context, deviceId int) (string, error)
	GetDeviceGroupNameById(ctx context.Context, groupId int) (string, error)
}

type DeviceGatewayService interface {
	Add(data DeviceData)
	Start(ctx context.Context) error
	Stop()
	UpdateDeviceLastSeen(ctx context.Context, deviceId int) error
	GetDeviceIdByName(ctx context.Context, deviceName string) (int, error)
	GetDeviceIdByGroupName(ctx context.Context, groupName string) ([]DeviceGroupData, error)
}

type DeviceGatewayClient interface {
	GetDeviceStatus(ctx context.Context, deviceId int) (bool, error)
	ExecuteCommand(ctx context.Context, req GatewayCommand) error
}
