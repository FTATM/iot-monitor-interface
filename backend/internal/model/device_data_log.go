package model

import "time"

type DeviceDataLog struct {
	Id         int64     `json:"Id" db:"id"`
	DeviceId   int       `json:"deviceId" db:"device_id"`
	ValueData  int       `json:"valueData" db:"value_data"`
	ReceivedAt time.Time `json:"receivedAt" db:"received_at"`
}

type DeviceDataLogReport struct {
	DeviceId   int       `json:"deviceId" db:"device_id"`
	ValueData  float64   `json:"valueData" db:"value_data"`
	ReceivedAt time.Time `json:"receivedAt" db:"received_at"`
	DeviceName string    `json:"deviceName" db:"device_name"`
}
