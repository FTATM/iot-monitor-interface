package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/FTATM/iot-monitor-interface/internal/model"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type scheduleEngineService struct {
	scheduleEngineRepo model.ScheduleEngineRepository
	prefixError        string
	engine             gocron.Scheduler
	deviceRepo         model.DeviceRepository // ⚡ NEW: To fetch dynamic group members
	gatewayClient      model.DeviceGatewayClient
	jobRegistry        sync.Map // Maps DB UUID (string) to gocron UUID (uuid.UUID) -- jobId and schedulId not the same jobId for internal work relate is schedulId is key and jobId is value
}

func NewSchedulerEngineService(ser model.ScheduleEngineRepository, dr model.DeviceRepository, gc model.DeviceGatewayClient) (model.ScheduleEngineService, error) {
	engine, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}
	return &scheduleEngineService{
		scheduleEngineRepo: ser,
		deviceRepo:         dr,
		gatewayClient:      gc,
		engine:             engine,
		prefixError:        "scheduleEngineService",
	}, nil
}

func (s *scheduleEngineService) SyncJob(ctx context.Context, id string) error {
	const fname = "SyncJob"
	// 1. Fetch the newly saved data from Postgres
	sched, err := s.scheduleEngineRepo.GetById(ctx, id)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
		}
	}

	// 2. SAFETY CHECK: Only allow 'active' schedules to be synced into memory
	if sched.Status != "active" {
		// Just in case it was running, kill it.
		s.CancelJob(id)
		return fmt.Errorf("[%s]>[%s]: rejected sync for schedule %s: status is '%s'", s.prefixError, fname, id, sched.Status)
	}

	// Remove any old version of this job from memory before adding the new one
	wasRunning := s.CancelJob(id)
	if wasRunning {
		slog.WarnContext(ctx, "Removed previous version of job before syncing", slog.String("scheduleId", id))
	}

	// Add the new job to memory
	if err := s.scheduleJob(sched); err != nil {
		return fmt.Errorf("[%s]>[%s] failed to add job %s to memory engine: %w", s.prefixError, fname, id, err)
	}

	slog.InfoContext(ctx, "Synced job into memory", slog.String("scheduleId", id))
	return nil
}

func (s *scheduleEngineService) Start(ctx context.Context) error {
	const fname = "Start"
	// Load active jobs from Repo on startup
	jobs, err := s.scheduleEngineRepo.GetActiveSchedules(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s] failed to load active schedules during startup:  %w", s.prefixError, fname, err)
	}

	successCount := 0
	for _, job := range jobs {
		// Capture and handle the error from scheduleJob
		if err := s.scheduleJob(&job); err != nil {
			slog.ErrorContext(ctx, "Synced job into memory", slog.String("scheduleId", job.ScheduleId), slog.String("error", err.Error()))
			continue // Skip to the next job
		}
		successCount++
	}

	slog.InfoContext(ctx, fmt.Sprintf("[STARTUP] Successfully loaded %d out of %d active schedules", successCount, len(jobs)))

	// Start the gocron engine
	s.engine.Start()

	return nil
}

func (s *scheduleEngineService) Shutdown(ctx context.Context) error {
	const fname = "Stop"
	err := s.engine.ShutdownWithContext(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	return nil
}

// Add a job to the memory engine
func (s *scheduleEngineService) scheduleJob(sched *model.Schedule) error {
	const fname = "scheduleJob"
	var jobDef gocron.JobDefinition
	if sched.ScheduleType == "recurring" && sched.CronExpression != nil {
		jobDef = gocron.CronJob(*sched.CronExpression, false)
	} else {
		jobDef = gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(sched.StartTime))
	}

	job, err := s.engine.NewJob(jobDef, gocron.NewTask(func() {}))
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	s.jobRegistry.Store(sched.ScheduleId, job.ID())
	s.engine.Update(job.ID(), jobDef, gocron.NewTask(s.executeDeviceTask, context.Background(), job.ID(), sched))

	return nil
}

// Cancel a job in memory
func (s *scheduleEngineService) CancelJob(schedId string) bool {
	val, ok := s.jobRegistry.Load(schedId)
	if ok {
		jobID := val.(uuid.UUID)
		s.engine.RemoveJob(jobID)
		s.jobRegistry.Delete(schedId)
		slog.Info("Cancelled job in memory", slog.String("scheduleId", schedId))
	}
	return ok
}

// The actual task that fires
func (s *scheduleEngineService) executeDeviceTask(ctx context.Context, jobID uuid.UUID, sched *model.Schedule) {
	slog.Info("Executing scheduled task", slog.String("scheduleId", sched.ScheduleId))

	// 1. Check completion status
	isFinished := (sched.ScheduleType == "one_time") || (sched.EndTime != nil && time.Now().After(*sched.EndTime))

	if isFinished {
		s.engine.RemoveJob(jobID)
		s.jobRegistry.Delete(sched.ScheduleId)
		s.scheduleEngineRepo.UpdateStatus(ctx, sched.ScheduleId, "completed")
		s.scheduleEngineRepo.UpdateLastRun(ctx, sched.ScheduleId)
		slog.Info("Job completed.", slog.String("scheduleId", sched.ScheduleId))
	} else {
		s.scheduleEngineRepo.UpdateLastRun(ctx, sched.ScheduleId)
	}

	// 2. Parse the TaskActionPayload
	var actionPayload model.TaskActionPayload
	if len(sched.TaskAction) > 0 {
		if err := json.Unmarshal(sched.TaskAction, &actionPayload); err != nil {
			slog.ErrorContext(ctx, "Failed to parse task action", slog.String("error", err.Error()))
			return
		}
	}

	// 3. Resolve Target Device IDs Dynamically
	var targetDeviceIds []int
	isGroupTarget := false

	if sched.DeviceGroupId != nil && *sched.DeviceGroupId > 0 {
		isGroupTarget = true

		// ⚡ DYNAMIC SYNC: Fetch the current devices in this group right now
		groupMap, err := s.deviceRepo.GetDeviceIdByDeviceGroupIds(ctx, []int{*sched.DeviceGroupId})
		if err != nil {
			slog.ErrorContext(ctx, "Failed to fetch group devices", slog.String("error", err.Error()))
			return
		}
		targetDeviceIds = groupMap[*sched.DeviceGroupId]

	} else if sched.DeviceId != nil && *sched.DeviceId > 0 {
		targetDeviceIds = []int{*sched.DeviceId}
	}

	if len(targetDeviceIds) == 0 {
		slog.WarnContext(ctx, "No devices found for schedule, aborting execution", slog.String("scheduleId", sched.ScheduleId))
		return
	}

	// 4. Fetch full device details (Protocols & Group IDs)
	devices, err := s.deviceRepo.GetDeviceForCommandByIds(ctx, targetDeviceIds)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to fetch device info for scheduled command", slog.String("error", err.Error()))
		return
	}

	// 5. Bundle Commands
	finalCommands := model.BuildGatewayCommands(devices, actionPayload, isGroupTarget)

	// 6. Send all commands to the Gateway
	for _, cmd := range finalCommands {
		// Because gocron runs this inside a background worker, 'ctx' is already detached from any HTTP requests. Safe to fire!
		if err := s.gatewayClient.ExecuteCommand(ctx, cmd); err != nil {
			slog.ErrorContext(ctx, "Failed to execute scheduled command via Gateway", slog.String("error", err.Error()))
		}
	}
}
