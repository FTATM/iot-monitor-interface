package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/FTATM/iot-monitor-interface/internal/model"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

type scheduleEngineService struct {
	scheduleEngineRepo model.ScheduleEngineRepository
	prefixError        string
	engine             gocron.Scheduler
	jobRegistry        sync.Map // Maps DB UUID (string) to gocron UUID (uuid.UUID) -- jobId and schedulId not the same jobId for internal work relate is schedulId is key and jobId is value
}

func NewSchedulerEngineService(ser model.ScheduleEngineRepository) (model.ScheduleEngineService, error) {
	engine, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}
	return &scheduleEngineService{
		scheduleEngineRepo: ser,
		engine:             engine,
	}, nil
}

func (s *scheduleEngineService) SyncJob(ctx context.Context, id string) error {
	const fname = "SyncJob"
	// 1. Fetch the newly saved data from Postgres
	sched, err := s.scheduleEngineRepo.GetById(ctx, id)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
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
func (s *scheduleEngineService) executeDeviceTask(ctx context.Context, jobID uuid.UUID, sched model.Schedule) {
	slog.Info("Sending action to device ", slog.String("scheduleId", sched.ScheduleId))

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
}
