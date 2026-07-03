package task

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/plugin"
	"github.com/theo-henon/lcloud/internal/volume"
	"gorm.io/gorm"
)

type Service struct {
	db        *gorm.DB
	volumes   *volume.Service
	executor  *Executor
	scheduler *Scheduler
	locks     *RunLocks
	events    *plugin.Service
}

func NewService(
	db *gorm.DB,
	volumes *volume.Service,
	executor *Executor,
	events *plugin.Service,
) *Service {
	s := &Service{
		db:       db,
		volumes:  volumes,
		executor: executor,
		locks:    NewRunLocks(),
		events:   events,
	}
	return s
}

func (s *Service) SetScheduler(scheduler *Scheduler) {
	s.scheduler = scheduler
}

type CreateTaskInput struct {
	Name         string
	Macro        string
	Scope        string
	VolumeID     *uuid.UUID
	Parameters   Parameters
	ScheduleType string
	Schedule     string
	Enabled      bool
}

type PatchTaskInput struct {
	Name         *string
	Macro        *string
	Scope        *string
	VolumeID     *uuid.UUID
	Parameters   Parameters
	ScheduleType *string
	Schedule     *string
	Enabled      *bool
}

type TaskView struct {
	Task
	OwnerEmail          string `json:"owner_email,omitempty"`
	VolumeName          string `json:"volume_name,omitempty"`
	ScheduleDescription string `json:"schedule_description"`
	LastRunStatus       string `json:"last_run_status,omitempty"`
}

func (s *Service) Create(ctx context.Context, claims *auth.Claims, input CreateTaskInput) (*TaskView, error) {
	if err := s.validateTaskInput(claims, input.Macro, input.Scope, input.VolumeID, input.Parameters, input.ScheduleType, input.Schedule); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	task := Task{
		ID:           uuid.New(),
		OwnerID:      claims.UserID,
		Name:         input.Name,
		Macro:        input.Macro,
		Scope:        input.Scope,
		VolumeID:     input.VolumeID,
		Parameters:   input.Parameters,
		ScheduleType: input.ScheduleType,
		Schedule:     input.Schedule,
		Enabled:      input.Enabled,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if task.Parameters == nil {
		task.Parameters = Parameters{}
	}

	next, err := NextRunAt(task.ScheduleType, task.Schedule, now)
	if err != nil {
		return nil, err
	}
	task.NextRunAt = next

	if err := s.db.Create(&task).Error; err != nil {
		return nil, err
	}
	// GORM Create skips false booleans; persist explicit disabled state.
	if !input.Enabled {
		if err := s.db.Model(&task).Update("enabled", false).Error; err != nil {
			_ = s.db.Delete(&Task{}, "id = ?", task.ID).Error
			return nil, err
		}
		task.Enabled = false
	}
	if s.scheduler != nil {
		if err := s.scheduler.Register(ctx, task); err != nil {
			_ = s.db.Delete(&Task{}, "id = ?", task.ID).Error
			return nil, err
		}
	}
	return s.toView(&task)
}

func (s *Service) List(claims *auth.Claims, volumeID *uuid.UUID, ownerID *uuid.UUID) ([]TaskView, error) {
	query := s.db.Order("created_at desc")
	if claims.Role != auth.RoleAdmin {
		query = query.Where("owner_id = ?", claims.UserID)
	} else if ownerID != nil {
		query = query.Where("owner_id = ?", *ownerID)
	}
	if volumeID != nil {
		query = query.Where("volume_id = ?", *volumeID)
	}

	var tasks []Task
	if err := query.Find(&tasks).Error; err != nil {
		return nil, err
	}

	views := make([]TaskView, 0, len(tasks))
	for i := range tasks {
		view, err := s.toView(&tasks[i])
		if err != nil {
			return nil, err
		}
		views = append(views, *view)
	}

	if claims.Role == auth.RoleAdmin && len(views) > 0 {
		ownerIDs := make([]uuid.UUID, len(views))
		for i := range views {
			ownerIDs[i] = views[i].OwnerID
		}
		emails, err := s.ownerEmailsByIDs(ownerIDs)
		if err != nil {
			return nil, err
		}
		for i := range views {
			views[i].OwnerEmail = emails[views[i].OwnerID]
		}
	}

	return views, nil
}

func (s *Service) Get(claims *auth.Claims, id uuid.UUID) (*TaskView, error) {
	task, err := s.findTask(id)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeTask(claims, task); err != nil {
		return nil, err
	}
	view, err := s.toView(task)
	if err != nil {
		return nil, err
	}
	if claims.Role == auth.RoleAdmin {
		emails, err := s.ownerEmailsByIDs([]uuid.UUID{task.OwnerID})
		if err != nil {
			return nil, err
		}
		view.OwnerEmail = emails[task.OwnerID]
	}
	return view, nil
}

func (s *Service) Patch(ctx context.Context, claims *auth.Claims, id uuid.UUID, input PatchTaskInput) (*TaskView, error) {
	task, err := s.findTask(id)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeTask(claims, task); err != nil {
		return nil, err
	}

	previous := *task
	previous.Parameters = cloneParams(task.Parameters)

	if input.Name != nil {
		task.Name = *input.Name
	}
	if input.Macro != nil {
		task.Macro = *input.Macro
	}
	if input.Scope != nil {
		task.Scope = *input.Scope
	}
	if input.VolumeID != nil {
		task.VolumeID = input.VolumeID
	}
	if input.Parameters != nil {
		task.Parameters = input.Parameters
	}
	if input.ScheduleType != nil {
		task.ScheduleType = *input.ScheduleType
	}
	if input.Schedule != nil {
		task.Schedule = *input.Schedule
	}
	if input.Enabled != nil {
		task.Enabled = *input.Enabled
	}
	task.UpdatedAt = time.Now().UTC()

	if err := s.validateTaskInput(claims, task.Macro, task.Scope, task.VolumeID, task.Parameters, task.ScheduleType, task.Schedule); err != nil {
		return nil, err
	}

	next, err := NextRunAt(task.ScheduleType, task.Schedule, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	task.NextRunAt = next

	if err := s.db.Save(task).Error; err != nil {
		return nil, err
	}
	if s.scheduler != nil {
		if err := s.scheduler.Register(ctx, *task); err != nil {
			if saveErr := s.db.Save(&previous).Error; saveErr != nil {
				return nil, fmt.Errorf("scheduler register failed: %w; rollback failed: %v", err, saveErr)
			}
			if regErr := s.scheduler.Register(ctx, previous); regErr != nil {
				return nil, fmt.Errorf("scheduler register failed: %w; rollback re-register failed: %v", err, regErr)
			}
			return nil, err
		}
	}
	return s.toView(task)
}

func (s *Service) Delete(ctx context.Context, claims *auth.Claims, id uuid.UUID) error {
	task, err := s.findTask(id)
	if err != nil {
		return err
	}
	if err := s.authorizeTask(claims, task); err != nil {
		return err
	}
	s.locks.CancelRun(id)
	if s.scheduler != nil {
		s.scheduler.Unregister(id)
	}
	return s.db.Delete(&Task{}, "id = ?", id).Error
}

func (s *Service) RunNow(ctx context.Context, claims *auth.Claims, id uuid.UUID, dryRunOverride bool) (uuid.UUID, error) {
	task, err := s.findTask(id)
	if err != nil {
		return uuid.Nil, err
	}
	if err := s.authorizeTask(claims, task); err != nil {
		return uuid.Nil, err
	}
	if !task.Enabled {
		return uuid.Nil, ErrTaskDisabled
	}

	runID := uuid.New()
	go s.executeTaskWithRunID(context.Background(), id, dryRunOverride, runID)
	return runID, nil
}

func (s *Service) ListRuns(claims *auth.Claims, taskID uuid.UUID, limit int) ([]TaskRun, error) {
	task, err := s.findTask(taskID)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeTask(claims, task); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var runs []TaskRun
	if err := s.db.Where("task_id = ?", taskID).Order("started_at desc").Limit(limit).Find(&runs).Error; err != nil {
		return nil, err
	}
	return runs, nil
}

func (s *Service) executeTask(ctx context.Context, taskID uuid.UUID, dryRunOverride bool) {
	s.executeTaskWithRunID(ctx, taskID, dryRunOverride, uuid.New())
}

func (s *Service) executeTaskWithRunID(ctx context.Context, taskID uuid.UUID, dryRunOverride bool, runID uuid.UUID) {
	if !s.locks.TryAcquireTask(taskID) {
		s.recordSkippedRun(taskID, runID, "task already running")
		return
	}
	defer s.locks.ReleaseTask(taskID)

	task, err := s.findTask(taskID)
	if err != nil || !task.Enabled {
		return
	}

	runCtx, cancel := context.WithCancel(ctx)
	s.locks.RegisterRunCancel(taskID, cancel)
	defer func() {
		cancel()
		s.locks.UnregisterRunCancel(taskID)
	}()

	params := task.Parameters
	if params == nil {
		params = Parameters{}
	}
	dryRun := IsDryRun(params) || dryRunOverride
	needsVolumeLock := task.Scope == ScopeVolume && IsDestructiveMacro(task.Macro) && !dryRun

	var volumeID uuid.UUID
	if task.VolumeID != nil {
		volumeID = *task.VolumeID
	}
	if needsVolumeLock {
		if !s.locks.TryAcquireVolume(volumeID) {
			s.recordSkippedRun(task.ID, runID, "volume busy")
			return
		}
		defer s.locks.ReleaseVolume(volumeID)
	}

	started := time.Now().UTC()
	run := TaskRun{
		ID:        runID,
		TaskID:    task.ID,
		Status:    RunStatusSuccess,
		StartedAt: started,
	}

	result, err := s.executor.Run(runCtx, task, dryRun)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			run.Status = RunStatusSkipped
			run.Message = "task cancelled"
		} else {
			run.Status = RunStatusFailed
			run.Error = err.Error()
			run.Message = "Task execution failed"
			s.emitTaskFailed(*task, err.Error())
		}
		s.finishRun(task, &run, started)
		return
	}

	run.AffectedCount = result.AffectedCount
	run.Message = result.Message
	if dryRun {
		run.Status = RunStatusDryRun
	}
	s.finishRun(task, &run, started)
	s.emitTaskExecuted(*task, result.AffectedCount, run.DurationMs, dryRun)
}

func (s *Service) finishRun(task *Task, run *TaskRun, started time.Time) {
	finished := time.Now().UTC()
	run.FinishedAt = finished
	run.DurationMs = finished.Sub(started).Milliseconds()

	if err := s.db.Create(run).Error; err == nil {
		s.purgeOldRuns(task.ID)
	}

	lastRun := finished
	task.LastRunAt = &lastRun
	next, err := NextRunAt(task.ScheduleType, task.Schedule, finished)
	if err == nil {
		task.NextRunAt = next
	}
	_ = s.db.Model(task).Updates(map[string]any{
		"last_run_at": lastRun,
		"next_run_at": task.NextRunAt,
		"updated_at":  finished,
	}).Error
}

func (s *Service) recordSkippedRun(taskID, runID uuid.UUID, message string) {
	now := time.Now().UTC()
	run := TaskRun{
		ID:        runID,
		TaskID:    taskID,
		Status:    RunStatusSkipped,
		Message:   message,
		StartedAt: now,
		FinishedAt: now,
	}
	_ = s.db.Create(&run).Error
}

func (s *Service) purgeOldRuns(taskID uuid.UUID) {
	var count int64
	s.db.Model(&TaskRun{}).Where("task_id = ?", taskID).Count(&count)
	if count <= maxRunsPerTask {
		return
	}
	excess := int(count) - maxRunsPerTask
	var old []TaskRun
	s.db.Where("task_id = ?", taskID).Order("started_at asc").Limit(excess).Find(&old)
	for _, run := range old {
		s.db.Delete(&run)
	}
}

func (s *Service) updateNextRunAt(taskID uuid.UUID, next *time.Time) error {
	return s.db.Model(&Task{}).Where("id = ?", taskID).Update("next_run_at", next).Error
}

func (s *Service) emitTaskExecuted(task Task, affected int, durationMs int64, dryRun bool) {
	if s.events == nil || dryRun {
		return
	}
	volID := ""
	if task.VolumeID != nil {
		volID = task.VolumeID.String()
	}
	s.events.PublishEvent(context.Background(), plugin.FromTaskExecuted(
		task.ID.String(), task.Macro, volID, affected, durationMs,
	))
}

func (s *Service) emitTaskFailed(task Task, errMsg string) {
	if s.events == nil {
		return
	}
	volID := ""
	if task.VolumeID != nil {
		volID = task.VolumeID.String()
	}
	s.events.PublishEvent(context.Background(), plugin.FromTaskFailed(
		task.ID.String(), task.Macro, volID, errMsg,
	))
}

func (s *Service) validateTaskInput(
	claims *auth.Claims,
	macro, scope string,
	volumeID *uuid.UUID,
	params Parameters,
	scheduleType, schedule string,
) error {
	if err := ValidateMacroScope(macro, scope); err != nil {
		return err
	}
	if err := ValidateParameters(macro, params); err != nil {
		return err
	}
	if err := ValidateSchedule(scheduleType, schedule); err != nil {
		return err
	}
	if scope == ScopeGlobal {
		if claims.Role != auth.RoleAdmin {
			return ErrGlobalAdminOnly
		}
		if volumeID != nil {
			return fmt.Errorf("%w: global tasks cannot have volume_id", ErrInvalidParameters)
		}
		return nil
	}
	if volumeID == nil {
		return ErrVolumeRequired
	}
	vol, err := s.volumes.GetByID(*volumeID)
	if err != nil {
		return ErrVolumeNotFound
	}
	if claims.Role != auth.RoleAdmin && vol.OwnerID != claims.UserID {
		return ErrForbidden
	}
	return nil
}

func (s *Service) findTask(id uuid.UUID) (*Task, error) {
	var task Task
	if err := s.db.First(&task, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}
	return &task, nil
}

func (s *Service) authorizeTask(claims *auth.Claims, task *Task) error {
	if claims.Role == auth.RoleAdmin {
		return nil
	}
	if task.OwnerID != claims.UserID {
		return ErrForbidden
	}
	return nil
}

func (s *Service) toView(task *Task) (*TaskView, error) {
	view := TaskView{
		Task:                *task,
		ScheduleDescription: DescribeSchedule(task.ScheduleType, task.Schedule),
	}
	if task.VolumeID != nil {
		vol, err := s.volumes.GetByID(*task.VolumeID)
		if err == nil {
			view.VolumeName = vol.Name
		}
	}
	var lastRun TaskRun
	if err := s.db.Where("task_id = ?", task.ID).Order("started_at desc").First(&lastRun).Error; err == nil {
		view.LastRunStatus = lastRun.Status
	}
	return &view, nil
}

func (s *Service) ownerEmailsByIDs(ids []uuid.UUID) (map[uuid.UUID]string, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]string{}, nil
	}

	unique := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		unique[id] = struct{}{}
	}
	deduped := make([]uuid.UUID, 0, len(unique))
	for id := range unique {
		deduped = append(deduped, id)
	}

	var rows []struct {
		ID    uuid.UUID
		Email string
	}
	if err := s.db.Table("users").Select("id, email").Where("id IN ?", deduped).Scan(&rows).Error; err != nil {
		return nil, err
	}

	out := make(map[uuid.UUID]string, len(rows))
	for _, row := range rows {
		out[row.ID] = row.Email
	}
	return out, nil
}

func (s *Service) ListEnabled() ([]Task, error) {
	var tasks []Task
	if err := s.db.Where("enabled = ?", true).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}
