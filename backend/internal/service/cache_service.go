package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/FTATM/iot-monitor-interface/internal/model"
)

const cacheTTL = 5 * time.Minute
const cleanupInterval = 10 * time.Minute

type cacheItem struct {
	data      any
	expiresAt time.Time
}

type cacheService struct {
	repo                 model.DeviceGatewayRepository
	deviceNameCache      sync.Map
	deviceGroupNameCache sync.Map
}

func NewCacheService(repo model.DeviceGatewayRepository) model.CacheService {
	return &cacheService{
		repo: repo,
	}
}

func (s *cacheService) GetDeviceIdByName(ctx context.Context, deviceName string) (int, error) {
	// 1. Check Cache
	if item, ok := s.deviceNameCache.Load(deviceName); ok {
		cached := item.(cacheItem)
		if time.Now().Before(cached.expiresAt) {
			return cached.data.(int), nil // HIT
		}
		s.deviceNameCache.Delete(deviceName) // EXPIRED
	}

	// 2. Cache MISS -> Fetch from DB
	dbId, err := s.repo.GetDeviceIdByName(ctx, deviceName)
	if err != nil || dbId <= 0 {
		return 0, fmt.Errorf("device not found: %w", err)
	}

	// 3. Store in Cache
	s.deviceNameCache.Store(deviceName, cacheItem{
		data:      dbId,
		expiresAt: time.Now().Add(cacheTTL),
	})

	return dbId, nil
}

func (s *cacheService) GetDeviceNamesByGroupName(ctx context.Context, groupName string) ([]string, error) {
	// ตรวจสอบใน Cache
	if item, ok := s.deviceGroupNameCache.Load(groupName); ok {
		cached := item.(cacheItem)
		if time.Now().Before(cached.expiresAt) {
			return cached.data.([]string), nil
		}
		s.deviceGroupNameCache.Delete(groupName)
	}

	// หากไม่พบ ให้ดึงจาก DB
	groupDataList, err := s.repo.GetDeviceIdByGroupName(ctx, groupName)
	if err != nil || len(groupDataList) == 0 {
		return nil, fmt.Errorf("group not found or empty")
	}

	deviceNames := groupDataList[0].DeviceNames

	// เก็บลง Cache
	s.deviceGroupNameCache.Store(groupName, cacheItem{
		data:      deviceNames,
		expiresAt: time.Now().Add(5 * time.Minute), // หรือใช้ cacheTTL ที่คุณกำหนดไว้
	})

	return deviceNames, nil
}

func (s *cacheService) InvalidateDevice(deviceName string) {
	s.deviceNameCache.Delete(deviceName)
}

func (s *cacheService) InvalidateGroup(groupName string) {
	s.deviceGroupNameCache.Delete(groupName)
}

// StartSweeper replaces the global sweeper you had in the listener package
func (s *cacheService) StartSweeper(ctx context.Context) {
	ticker := time.NewTicker(cleanupInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.cleanMap(&s.deviceNameCache, "deviceNameCache")
				s.cleanMap(&s.deviceGroupNameCache, "deviceGroupNameCache")
			case <-ctx.Done():
				ticker.Stop()
				slog.Info("Shutting down cache sweeper...")
				return
			}
		}
	}()
}

func (s *cacheService) cleanMap(cache *sync.Map, cacheName string) {
	now := time.Now()
	expiredCount := 0

	cache.Range(func(key, value any) bool {
		cached := value.(cacheItem)
		if now.After(cached.expiresAt) {
			cache.Delete(key)
			expiredCount++
		}
		return true
	})

	if expiredCount > 0 {
		slog.Debug("Cache swept", slog.String("cache", cacheName), slog.Int("cleared", expiredCount))
	}
}
