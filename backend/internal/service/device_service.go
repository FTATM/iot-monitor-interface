package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/FTATM/iot-monitor-interface/internal/model"
)

type deviceService struct {
	txManager    model.TransactionManager
	prefixError  string
	deviceRepo   model.DeviceRepository
	auditLogRepo model.AuditLogRepository
	clientsMap   map[string][]chan model.ChartData
	mutex        sync.RWMutex
}

func NewDeviceService(txManager model.TransactionManager, dr model.DeviceRepository, alog model.AuditLogRepository) model.DeviceService {
	return &deviceService{
		txManager:    txManager,
		prefixError:  "deviceService",
		deviceRepo:   dr,
		auditLogRepo: alog,
		clientsMap:   make(map[string][]chan model.ChartData),
		mutex:        sync.RWMutex{},
	}
}

func (s *deviceService) GetAllDeviceDetail(ctx context.Context) ([]model.DeviceDetail, error) {
	const fname = "GetAllDeviceDetail"
	devices, err := s.deviceRepo.GetAll(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	deviceDetails := make([]model.DeviceDetail, 0, len(devices))

	for _, d := range devices {
		detail := model.DeviceDetail{
			DeviceId:    d.DeviceId,
			DeviceName:  d.DeviceName,
			Protocol:    d.Protocol,
			ValueData:   d.ValueData,
			IsActive:    d.IsActive,
			IsConnected: false,
			LastSeenAt:  d.LastSeenAt,
		}
		deviceDetails = append(deviceDetails, detail)
	}

	return deviceDetails, nil
}

func (s *deviceService) CreateDevice(ctx context.Context, createDevice []model.DeviceCreate, authUserId int) error {
	const fname = "CreateDevice"
	var err error
	if len(createDevice) == 0 {
		return nil
	}

	devices := make([]model.Device, 0, len(createDevice))
	for _, createReq := range createDevice {
		device := model.Device{
			DeviceName: createReq.DeviceName,
			Protocol:   createReq.Protocol,
			IsActive:   createReq.IsActive,
		}
		devices = append(devices, device)
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	if err = s.deviceRepo.Create(tx.Context(), devices); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	auditlogs := make([]model.AuditLog, 0, len(devices))
	for _, d := range devices {

		newData, err := model.StructToDynamicJSON(d)
		if err != nil {
			return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
		}
		audit := model.AuditLog{
			EntityType: "device",
			EntityId:   strconv.Itoa(d.DeviceId),
			Action:     model.CreateAction,
			ChangedBy:  authUserId,
			OldData:    nil,
			NewData:    newData,
		}
		auditlogs = append(auditlogs, audit)
	}

	if err = s.auditLogRepo.Create(tx.Context(), auditlogs); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil
}

func (s *deviceService) UpdateDevice(ctx context.Context, updateDevice *model.DeviceUpdate, authUserId int) error {
	const fname = "UpdateDevice"
	var err error

	device := model.Device{
		DeviceId: updateDevice.DeviceId,
		IsActive: updateDevice.IsActive,
		Protocol: updateDevice.Protocol,
	}

	oldDevice, err := s.deviceRepo.GetById(ctx, device.DeviceId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	// check same then return
	if oldDevice.IsSame(device) {
		return nil
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	if err = s.deviceRepo.Update(tx.Context(), &device); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	auditlogs := make([]model.AuditLog, 0, 1)
	oldData, err := model.StructToDynamicJSON(oldDevice)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	newData, err := model.StructToDynamicJSON(device)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	audit := model.AuditLog{
		EntityType: "device",
		EntityId:   strconv.Itoa(device.DeviceId),
		Action:     model.UpdateAction,
		ChangedBy:  authUserId,
		OldData:    oldData,
		NewData:    newData,
	}
	auditlogs = append(auditlogs, audit)

	if err = s.auditLogRepo.Create(tx.Context(), auditlogs); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil
}

func (s *deviceService) DeleteDevice(ctx context.Context, deviceId, authUserId int) error {
	const fname = "DeleteDevice"
	var err error

	oldDevice, err := s.deviceRepo.GetById(ctx, deviceId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	if err = s.deviceRepo.Delete(tx.Context(), deviceId); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	auditlogs := make([]model.AuditLog, 0, 1)
	oldData, err := model.StructToDynamicJSON(oldDevice)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	audit := model.AuditLog{
		EntityType: "device",
		EntityId:   strconv.Itoa(deviceId),
		Action:     model.DeleteAction,
		ChangedBy:  authUserId,
		OldData:    oldData,
		NewData:    nil,
	}
	auditlogs = append(auditlogs, audit)

	if err = s.auditLogRepo.Create(tx.Context(), auditlogs); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil
}

func (s *deviceService) GetProtocolType(ctx context.Context) ([]string, error) {
	const fname = "GetProtocolType"
	protocols, err := s.deviceRepo.GetProtocolType(ctx)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return protocols, nil
}

func (s *deviceService) AddClient(deviceId string, clientChan chan model.ChartData) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.clientsMap[deviceId] = append(s.clientsMap[deviceId], clientChan)
}

func (s *deviceService) RemoveClient(deviceId string, clientChan chan model.ChartData) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Find and remove the closed channel from the slice
	clients := s.clientsMap[deviceId]
	for i, ch := range clients {
		if ch == clientChan {
			s.clientsMap[deviceId] = append(clients[:i], clients[i+1:]...)
			break
		}
	}
}

// Update StartPublic to use the struct's state
func (s *deviceService) StartPublic(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			s.mutex.RLock()

			// STEP 1: Find all unique devices that ANY widget is watching right now
			activeDeviceIDs := make(map[int]bool)
			for requestedIDs, clients := range s.clientsMap {
				if len(clients) == 0 {
					continue
				}
				for idStr := range strings.SplitSeq(requestedIDs, ",") {
					if idInt, err := strconv.Atoi(idStr); err == nil {
						activeDeviceIDs[idInt] = true
					}
				}
			}

			// STEP 2: Fetch data for those unique devices EXACTLY ONCE
			masterDataMap := make(map[int]model.ChartDeviceData)
			for deviceId := range activeDeviceIDs {

				// ⚡ DB CALL HAPPENS HERE (Guaranteed only once per device!) ⚡
				chartDeviceData, err := s.deviceRepo.GetByIdChartDeviceData(ctx, deviceId)
				if err != nil {
					slog.DebugContext(ctx, "Failed to fetch chart data for device",
						slog.Int("deviceId", deviceId),
						slog.String("error", err.Error()))
					continue
				}
				masterDataMap[deviceId] = chartDeviceData
			}

			// STEP 3: Distribute the data back to the specific widgets
			for requestedIDs, clients := range s.clientsMap {
				if len(clients) == 0 {
					continue
				}

				// Build a custom payload for this specific group of widgets
				clientPayloadMap := make(map[int]model.ChartDeviceData)
				for idStr := range strings.SplitSeq(requestedIDs, ",") {
					if idInt, err := strconv.Atoi(idStr); err == nil {
						clientPayloadMap[idInt] = masterDataMap[idInt]
					}
				}

				payload := model.ChartData{
					DeviceData: clientPayloadMap,
				}

				// Broadcast to every widget in this group
				for _, ch := range clients {
					select {
					case ch <- payload:
					default:
					}
				}
			}

			s.mutex.RUnlock()
		}
	}
}

func (s *deviceService) GetAllDeviceName(ctx context.Context) ([]model.DeviceDetail, error) {
	const fname = "GetAllDeviceName"
	devices, err := s.deviceRepo.GetAll(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	deviceDetails := make([]model.DeviceDetail, 0, len(devices))

	for _, d := range devices {
		detail := model.DeviceDetail{
			DeviceId:   d.DeviceId,
			DeviceName: d.DeviceName,
			IsActive:   d.IsActive,
		}
		deviceDetails = append(deviceDetails, detail)
	}

	return deviceDetails, nil
}

func (s *deviceService) GetChartHistory(ctx context.Context, deviceIds []int, maxPoints int, fromTime, toTime time.Time) (map[int][][2]float64, error) {
	count, err := s.deviceRepo.CountData(ctx, deviceIds, fromTime, toTime)
	if err != nil {
		return nil, err
	}

	var logs []model.DeviceDataLog

	// 2. ⚡ DYNAMIC DECISION LOGIC ⚡
	if count <= maxPoints {
		// Scenario A: Safe to send raw data!
		// The user zoomed in, or the device hasn't sent many points yet.
		logs, err = s.deviceRepo.GetRawData(ctx, deviceIds, fromTime, toTime, maxPoints)
		if err != nil {
			return nil, err
		}
	} else {
		// Scenario B: Too much data!
		// We must aggregate it so we don't crash the frontend.
		totalDuration := toTime.Sub(fromTime)

		// Calculate the dynamic bucket size
		bucketDuration := max(totalDuration/time.Duration(maxPoints), time.Second)
		bucketInterval := fmt.Sprintf("%f seconds", bucketDuration.Seconds())

		logs, err = s.deviceRepo.GetAggregatedData(ctx, deviceIds, fromTime, toTime, bucketInterval)
		if err != nil {
			return nil, err
		}
	}

	// 3. Map the chosen data perfectly for ECharts
	historyData := make(map[int][][2]float64)

	for _, log := range logs {
		tsMillis := float64(log.ReceivedAt.UnixMilli())
		valData := float64(log.ValueData)

		historyData[log.DeviceId] = append(historyData[log.DeviceId], [2]float64{tsMillis, valData})
	}

	return historyData, nil
}
