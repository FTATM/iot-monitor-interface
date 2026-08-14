package model

import "time"

type DeviceDataLog struct {
	Id         int64     `json:"Id" db:"Id"`
	DeviceId   int       `json:"deviceId" db:"device_id"`
	ValueData  int       `json:"valueData" db:"value_data"`
	Source     string    `json:"source" db:"source"`
	ReceivedAt time.Time `json:"receivedAt" db:"received_at"`
}
