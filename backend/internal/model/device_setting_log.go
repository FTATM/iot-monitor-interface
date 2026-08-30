package model

type DeviceSettingLog struct {
	LogId      int    `json:"logId" db:"log_id"`
	DeviceId   int    `json:"deviceId" db:"device_id"`
	DeviceName string `json:"deviceName" db:"device_name"`
	Active     bool   `json:"active" db:"active"`
}
