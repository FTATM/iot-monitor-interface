package model

import "context"

type CacheService interface {
	GetDeviceIdByName(ctx context.Context, deviceName string) (int, error)
	GetDeviceNamesByGroupName(ctx context.Context, groupName string) ([]string, error)
	InvalidateDevice(deviceName string)
	InvalidateGroup(groupName string)
	StartSweeper(ctx context.Context) // Replaces your background cache sweeper
}
