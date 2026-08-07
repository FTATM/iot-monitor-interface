package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/FTATM/iot-monitor-interface/internal/model"
)

type deviceService struct {
	txManager    model.TransactionManager
	prefixError  string
	deviceRepo   model.DeviceRepository
	auditLogRepo model.AuditLogRepository
}

func NewDeviceService(txManager model.TransactionManager, dr model.DeviceRepository, alog model.AuditLogRepository) model.DeviceService {
	return &deviceService{txManager: txManager, prefixError: "deviceService", deviceRepo: dr, auditLogRepo: alog}
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
			ValueData:   d.ValueData,
			IsActive:    d.IsActive,
			IsConnected: d.IsConnected,
			LastSeenAt:  d.LastSeenAt,
		}
		deviceDetails = append(deviceDetails, detail)
	}

	return deviceDetails, nil
}

func (s *deviceService) CreateDevice(ctx context.Context, createDeviceReq []model.DeviceCreate, authUserId int) error {
	const fname = "CreateDevice"
	var err error
	if len(createDeviceReq) == 0 {
		return nil
	}

	createDevices := make([]model.Device, 0, len(createDeviceReq))
	for _, createReq := range createDeviceReq {
		device := model.Device{
			DeviceName: createReq.DeviceName,
			IsActive:   createReq.IsActive,
		}
		createDevices = append(createDevices, device)
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	if err = s.deviceRepo.Create(tx.Context(), createDevices); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	auditlogs := make([]model.AuditLog, 0, len(createDevices))
	for _, d := range createDevices {

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

func (s *deviceService) UpdateDevice(ctx context.Context, updateDeviceReq *model.DeviceUpdate, authUserId int) error {
	const fname = "UpdateDevice"
	var err error

	updateDevice := model.Device{
		DeviceId: updateDeviceReq.DeviceId,
		IsActive: updateDeviceReq.IsActive,
	}

	oldDevice, err := s.deviceRepo.GetById(ctx, updateDevice.DeviceId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	// check same then return
	if oldDevice.IsSame(updateDevice) {
		return nil
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	if err = s.deviceRepo.Update(tx.Context(), &updateDevice); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	auditlogs := make([]model.AuditLog, 0, 1)
	oldData, err := model.StructToDynamicJSON(oldDevice)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	newData, err := model.StructToDynamicJSON(updateDevice)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	audit := model.AuditLog{
		EntityType: "device",
		EntityId:   strconv.Itoa(updateDevice.DeviceId),
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
