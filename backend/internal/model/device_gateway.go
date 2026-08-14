package model

import (
	"context"
	"net"
	"time"
)

type DeviceDataRequest struct {
	DeviceId   int       `json:"deviceId" db:"device_id"`
	DeviceName string    `json:"deviceName,omitempty" db:"device_name"`
	ValueData  int       `json:"valueData" db:"value_data"`
	Source     string    `json:"source" db:"source"`
	ReceivedAt time.Time `json:"-"` // use in system
}

type CommandRequest struct {
	DeviceId int    `json:"deviceId"`
	Command  string `json:"command"`
	Protocol string `json:"protocol"`
}

// SessionManagerService defines the contract for managing active connections & command routing
type SessionManagerService interface {
	SetMQTTClient(client any)
	SetUDPServer(conn *net.UDPConn)
	RegisterTCP(deviceId int, conn net.Conn)
	UnregisterTCP(deviceId int, conn net.Conn)
	RegisterUDP(deviceId int, addr *net.UDPAddr)
	RouteCommand(ctx context.Context, req *CommandRequest) error
	PopHTTPCommand(deviceId int) (string, bool)
	MarkDeviceActive(deviceId int)
	IsDeviceOnline(deviceId int) bool
}

type DeviceGatewayRepository interface {
	BulkUpsertDeviceData(ctx context.Context, data []DeviceDataRequest) error
}

type DeviceGatewayService interface {
	Add(data DeviceDataRequest)
	Start(ctx context.Context) error
	Stop()
}

type DeviceGatewayClient interface {
	GetDeviceStatus(ctx context.Context, deviceId int) (bool, error)
	ExecuteManualCommand(ctx context.Context, req *CommandRequest) error
}
