package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/FTATM/iot-monitor-interface/internal/model"
)

type scheduleService struct {
	txManager    model.TransactionManager
	prefixError  string
	scheduleRepo model.ScheduleRepository
	auditLogRepo model.AuditLogRepository
}

func NewScheduleService(txManager model.TransactionManager, sr model.ScheduleRepository, alog model.AuditLogRepository) model.ScheduleService {
	return &scheduleService{txManager: txManager, prefixError: "scheduleService", scheduleRepo: sr, auditLogRepo: alog}
}

func (s *scheduleService) GetAllDetail(ctx context.Context) ([]model.ScheduleDetail, error) {
	const fname = "GetAllDetail"
	schedules, err := s.scheduleRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	scheduleDetails := make([]model.ScheduleDetail, 0, len(schedules))

	for _, s := range schedules {
		detail := model.ScheduleDetail{
			ScheduleId:     s.ScheduleId,
			DeviceId:       s.DeviceId,
			ScheduleType:   s.ScheduleType,
			Action:         s.Action,
			Status:         s.Status,
			StartTime:      s.StartTime,
			EndTime:        s.EndTime,
			CronExpression: s.CronExpression,
		}
		scheduleDetails = append(scheduleDetails, detail)
	}

	return scheduleDetails, nil

}
func (s *scheduleService) CreateSchedule(ctx context.Context, createdSched *model.CreateScheduleReq, userId int) error {
	const fname = "CreateSchedule"
	var err error

	sched := model.Schedule{
		DeviceId:       createdSched.DeviceId,
		Action:         createdSched.Action,
		ScheduleType:   createdSched.Action,
		Status:         createdSched.Status,
		StartTime:      createdSched.StartTime,
		EndTime:        createdSched.EndTime,
		CronExpression: createdSched.CronExpression,
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	if err = s.scheduleRepo.Create(ctx, &sched); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	auditlogs := make([]model.AuditLog, 0, 1)
	newData, err := model.StructToDynamicJSON(sched)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	audit := model.AuditLog{
		EntityType: "schedule",
		EntityId:   sched.ScheduleId,
		Action:     model.CreateAction,
		ChangedBy:  userId,
		OldData:    nil,
		NewData:    newData,
	}

	auditlogs = append(auditlogs, audit)

	if err = s.auditLogRepo.Create(ctx, auditlogs); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil

}
func (s *scheduleService) UpdateSchedule(ctx context.Context, updateSchedReq *model.UpdateScheduleReq, userId int) error {
	const fname = "UpdateSchedule"
	var err error

	sched := model.Schedule{
		ScheduleId:     updateSchedReq.ScheduleId,
		DeviceId:       updateSchedReq.DeviceId,
		Action:         updateSchedReq.Action,
		ScheduleType:   updateSchedReq.Action,
		Status:         updateSchedReq.Status,
		StartTime:      updateSchedReq.StartTime,
		EndTime:        updateSchedReq.EndTime,
		CronExpression: updateSchedReq.CronExpression,
	}

	oldSched, err := s.scheduleRepo.GetById(ctx, updateSchedReq.ScheduleId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	// check same then return
	if oldSched.IsSame(sched) {
		return nil
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	if err = s.scheduleRepo.Update(ctx, &sched); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	auditlogs := make([]model.AuditLog, 0, 1)
	oldData, err := model.StructToDynamicJSON(oldSched)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	newData, err := model.StructToDynamicJSON(sched)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	audit := model.AuditLog{
		EntityType: "schedule",
		EntityId:   strconv.Itoa(sched.DeviceId),
		Action:     model.UpdateAction,
		ChangedBy:  userId,
		OldData:    oldData,
		NewData:    newData,
	}
	auditlogs = append(auditlogs, audit)

	if err = s.auditLogRepo.Create(ctx, auditlogs); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil
}
