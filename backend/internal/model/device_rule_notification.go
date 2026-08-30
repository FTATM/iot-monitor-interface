package model

import "time"

type DeviceRuleNotification struct {
	RuleId          int        `json:"ruleId" db:"rule_id"`
	DeviceId        int        `json:"deviceId" db:"device_id"`
	Condition       string     `json:"condition" db:"condition"`
	Threshold       int        `json:"threshold" db:"threshold"`
	Reason          string     `json:"reason" db:"reason"`
	Active          bool       `json:"active" db:"active"`
	LastTriggeredAt *time.Time `json:"lastTriggeredAt,omitempty" db:"last_triggered_at"`
	CooldownMinutes int        `json:"cooldownMinutes,omitempty" db:"cooldown_minutes"`
	CreatedAt       time.Time  `json:"-" db:"created_at"`
	UpdatedAt       time.Time  `json:"-" db:"updated_at"`
}

type DeviceRuleNotificationDetail struct {
	DeviceRuleNotification
	DeviceName string `json:"deviceName" db:"device_name"`
}
