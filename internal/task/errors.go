package task

import "errors"

var (
	ErrTaskNotFound      = errors.New("task not found")
	ErrInvalidMacro      = errors.New("invalid macro")
	ErrInvalidParameters = errors.New("invalid parameters")
	ErrInvalidSchedule   = errors.New("invalid schedule")
	ErrVolumeRequired    = errors.New("volume required")
	ErrGlobalAdminOnly   = errors.New("global tasks require admin")
	ErrVolumeNotFound    = errors.New("volume not found")
	ErrForbidden         = errors.New("forbidden")
	ErrTaskAlreadyRunning = errors.New("task already running")
	ErrVolumeTaskBusy    = errors.New("volume task busy")
	ErrTaskDisabled      = errors.New("task disabled")
)
