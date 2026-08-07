package model

import "time"

type DeviceRequestLog struct {
	LogId     int        `json:"logId" db:"log_id"`
	DeviceId  int        `json:"deviceId" db:"device_id"`
	ValueData int        `json:"deviceValue" db:"value_data"`
	SentAt    *time.Time `json:"sentAt" db:"sent_at"`
}
