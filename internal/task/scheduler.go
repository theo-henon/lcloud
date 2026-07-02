package task

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

type Scheduler struct {
	cron    gocron.Scheduler
	service *Service
	mu      sync.Mutex
	jobs    map[uuid.UUID]uuid.UUID
}

func NewScheduler(service *Service) (*Scheduler, error) {
	cron, err := gocron.NewScheduler(gocron.WithLocation(time.UTC))
	if err != nil {
		return nil, err
	}
	return &Scheduler{
		cron:    cron,
		service: service,
		jobs:    make(map[uuid.UUID]uuid.UUID),
	}, nil
}

func (s *Scheduler) Start() {
	s.cron.Start()
}

func (s *Scheduler) Shutdown() error {
	return s.cron.Shutdown()
}

func (s *Scheduler) LoadAll(ctx context.Context, tasks []Task) error {
	for i := range tasks {
		if err := s.Register(ctx, tasks[i]); err != nil {
			return err
		}
	}
	return nil
}

func (s *Scheduler) Register(ctx context.Context, task Task) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()

	s.unregisterLocked(task.ID)
	if !task.Enabled {
		return s.updateNextRunLocked(task)
	}

	taskID := task.ID
	jobFn := func() {
		s.service.executeTask(context.Background(), taskID, false)
	}

	var job gocron.Job
	var err error
	switch task.ScheduleType {
	case ScheduleTypeCron:
		job, err = s.cron.NewJob(
			gocron.CronJob(task.Schedule, false),
			gocron.NewTask(jobFn),
		)
	case ScheduleTypeInterval:
		d, parseErr := time.ParseDuration(task.Schedule)
		if parseErr != nil {
			return parseErr
		}
		job, err = s.cron.NewJob(
			gocron.DurationJob(d),
			gocron.NewTask(jobFn),
		)
	default:
		return ErrInvalidSchedule
	}
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidSchedule, err)
	}

	s.jobs[task.ID] = job.ID()
	return s.updateNextRunLocked(task)
}

func (s *Scheduler) Unregister(taskID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.unregisterLocked(taskID)
}

func (s *Scheduler) unregisterLocked(taskID uuid.UUID) {
	jobID, ok := s.jobs[taskID]
	if !ok {
		return
	}
	_ = s.cron.RemoveJob(jobID)
	delete(s.jobs, taskID)
}

func (s *Scheduler) updateNextRunLocked(task Task) error {
	next, err := NextRunAt(task.ScheduleType, task.Schedule, time.Now().UTC())
	if err != nil {
		return err
	}
	return s.service.updateNextRunAt(task.ID, next)
}

func (s *Scheduler) NextRunFor(task Task) (*time.Time, error) {
	return NextRunAt(task.ScheduleType, task.Schedule, time.Now().UTC())
}
