package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/task"
	"github.com/theo-henon/lcloud/pkg/httputil"
)

type TaskHandler struct {
	tasks *task.Service
}

func NewTaskHandler(tasks *task.Service) *TaskHandler {
	return &TaskHandler{tasks: tasks}
}

type createTaskRequest struct {
	Name         string            `json:"name" binding:"required"`
	Macro        string            `json:"macro" binding:"required"`
	Scope        string            `json:"scope" binding:"required"`
	VolumeID     *string           `json:"volume_id"`
	Parameters   task.Parameters   `json:"parameters"`
	ScheduleType string            `json:"schedule_type" binding:"required"`
	Schedule     string            `json:"schedule" binding:"required"`
	Enabled      *bool             `json:"enabled"`
}

type patchTaskRequest struct {
	Name         *string         `json:"name"`
	Macro        *string         `json:"macro"`
	Scope        *string         `json:"scope"`
	VolumeID     *string         `json:"volume_id"`
	Parameters   task.Parameters `json:"parameters"`
	ScheduleType *string         `json:"schedule_type"`
	Schedule     *string         `json:"schedule"`
	Enabled      *bool           `json:"enabled"`
}

func (h *TaskHandler) List(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httputil.Unauthorized(c, "unauthorized")
		return
	}

	var volumeID *uuid.UUID
	if raw := c.Query("volume_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			httputil.BadRequest(c, "invalid volume_id")
			return
		}
		volumeID = &id
	}

	tasks, err := h.tasks.List(claims, volumeID)
	if err != nil {
		httputil.InternalError(c, "unable to list tasks")
		return
	}
	httputil.JSON(c, http.StatusOK, gin.H{"tasks": tasks})
}

func (h *TaskHandler) Create(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httputil.Unauthorized(c, "unauthorized")
		return
	}

	var req createTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "invalid request body")
		return
	}

	input := task.CreateTaskInput{
		Name:         req.Name,
		Macro:        req.Macro,
		Scope:        req.Scope,
		Parameters:   req.Parameters,
		ScheduleType: req.ScheduleType,
		Schedule:     req.Schedule,
		Enabled:      true,
	}
	if req.Enabled != nil {
		input.Enabled = *req.Enabled
	}
	if req.VolumeID != nil {
		id, err := uuid.Parse(*req.VolumeID)
		if err != nil {
			httputil.BadRequest(c, "invalid volume_id")
			return
		}
		input.VolumeID = &id
	}

	created, err := h.tasks.Create(c.Request.Context(), claims, input)
	if err != nil {
		h.handleTaskError(c, err)
		return
	}
	httputil.JSON(c, http.StatusCreated, created)
}

func (h *TaskHandler) Get(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httputil.Unauthorized(c, "unauthorized")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.BadRequest(c, "invalid task id")
		return
	}

	record, err := h.tasks.Get(claims, id)
	if err != nil {
		h.handleTaskError(c, err)
		return
	}
	httputil.JSON(c, http.StatusOK, record)
}

func (h *TaskHandler) Patch(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httputil.Unauthorized(c, "unauthorized")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.BadRequest(c, "invalid task id")
		return
	}

	var req patchTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "invalid request body")
		return
	}

	input := task.PatchTaskInput{Parameters: req.Parameters}
	if req.Name != nil {
		input.Name = req.Name
	}
	if req.Macro != nil {
		input.Macro = req.Macro
	}
	if req.Scope != nil {
		input.Scope = req.Scope
	}
	if req.VolumeID != nil {
		volID, err := uuid.Parse(*req.VolumeID)
		if err != nil {
			httputil.BadRequest(c, "invalid volume_id")
			return
		}
		input.VolumeID = &volID
	}
	if req.ScheduleType != nil {
		input.ScheduleType = req.ScheduleType
	}
	if req.Schedule != nil {
		input.Schedule = req.Schedule
	}
	if req.Enabled != nil {
		input.Enabled = req.Enabled
	}

	updated, err := h.tasks.Patch(c.Request.Context(), claims, id, input)
	if err != nil {
		h.handleTaskError(c, err)
		return
	}
	httputil.JSON(c, http.StatusOK, updated)
}

func (h *TaskHandler) Delete(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httputil.Unauthorized(c, "unauthorized")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.BadRequest(c, "invalid task id")
		return
	}

	if err := h.tasks.Delete(c.Request.Context(), claims, id); err != nil {
		h.handleTaskError(c, err)
		return
	}
	httputil.JSON(c, http.StatusOK, gin.H{"status": "ok"})
}

func (h *TaskHandler) Run(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httputil.Unauthorized(c, "unauthorized")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.BadRequest(c, "invalid task id")
		return
	}

	dryRun := c.Query("dry_run") == "true"
	runID, err := h.tasks.RunNow(c.Request.Context(), claims, id, dryRun)
	if err != nil {
		h.handleTaskError(c, err)
		return
	}
	httputil.JSON(c, http.StatusAccepted, gin.H{"run_id": runID.String()})
}

func (h *TaskHandler) Runs(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httputil.Unauthorized(c, "unauthorized")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.BadRequest(c, "invalid task id")
		return
	}

	limit := 20
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			httputil.BadRequest(c, "invalid limit")
			return
		}
		limit = parsed
	}

	runs, err := h.tasks.ListRuns(claims, id, limit)
	if err != nil {
		h.handleTaskError(c, err)
		return
	}
	httputil.JSON(c, http.StatusOK, gin.H{"runs": runs})
}

func (h *TaskHandler) handleTaskError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, task.ErrTaskNotFound):
		httputil.JSON(c, http.StatusNotFound, gin.H{"error": "TASK_NOT_FOUND"})
	case errors.Is(err, task.ErrInvalidMacro):
		httputil.Unprocessable(c, "INVALID_MACRO", err.Error())
	case errors.Is(err, task.ErrInvalidParameters):
		httputil.Unprocessable(c, "INVALID_PARAMETERS", err.Error())
	case errors.Is(err, task.ErrInvalidSchedule):
		httputil.Unprocessable(c, "INVALID_SCHEDULE", err.Error())
	case errors.Is(err, task.ErrVolumeRequired):
		httputil.Unprocessable(c, "VOLUME_REQUIRED", err.Error())
	case errors.Is(err, task.ErrGlobalAdminOnly):
		httputil.Forbidden(c, "global tasks require admin")
	case errors.Is(err, task.ErrVolumeNotFound):
		httputil.JSON(c, http.StatusNotFound, gin.H{"error": "VOLUME_NOT_FOUND"})
	case errors.Is(err, task.ErrForbidden):
		httputil.Forbidden(c, "forbidden")
	case errors.Is(err, task.ErrTaskAlreadyRunning):
		httputil.Unprocessable(c, "TASK_ALREADY_RUNNING", err.Error())
	case errors.Is(err, task.ErrVolumeTaskBusy):
		httputil.Unprocessable(c, "VOLUME_TASK_BUSY", err.Error())
	case errors.Is(err, task.ErrTaskDisabled):
		httputil.Unprocessable(c, "TASK_DISABLED", err.Error())
	default:
		httputil.InternalError(c, "task operation failed")
	}
}
