package task

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

const maxRunsPerTask = 200

type Parameters map[string]any

func (p Parameters) Value() (driver.Value, error) {
	if p == nil {
		return "{}", nil
	}
	return json.Marshal(p)
}

func (p *Parameters) Scan(value any) error {
	if value == nil {
		*p = Parameters{}
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, p)
	case string:
		return json.Unmarshal([]byte(v), p)
	default:
		return errors.New("unsupported Parameters scan type")
	}
}

const (
	ScopeVolume = "volume"
	ScopeGlobal = "global"

	ScheduleTypeCron     = "cron"
	ScheduleTypeInterval = "interval"

	RunStatusSuccess = "success"
	RunStatusFailed  = "failed"
	RunStatusSkipped = "skipped"
	RunStatusDryRun  = "dry_run"
)

type Task struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	OwnerID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"owner_id"`
	Name         string     `gorm:"not null" json:"name"`
	Macro        string     `gorm:"not null;index" json:"macro"`
	Scope        string     `gorm:"not null" json:"scope"`
	VolumeID     *uuid.UUID `gorm:"type:uuid;index" json:"volume_id,omitempty"`
	Parameters   Parameters `gorm:"type:jsonb;not null;default:'{}'" json:"parameters"`
	ScheduleType string     `gorm:"not null" json:"schedule_type"`
	Schedule     string     `gorm:"not null" json:"schedule"`
	Enabled      bool       `gorm:"not null;default:true" json:"enabled"`
	NextRunAt    *time.Time `gorm:"index" json:"next_run_at,omitempty"`
	LastRunAt    *time.Time `json:"last_run_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type TaskRun struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TaskID        uuid.UUID `gorm:"type:uuid;not null;index" json:"task_id"`
	Status        string    `gorm:"not null" json:"status"`
	AffectedCount int       `gorm:"not null;default:0" json:"affected_count"`
	Message       string    `json:"message"`
	Error         string    `json:"error,omitempty"`
	StartedAt     time.Time `gorm:"not null;index" json:"started_at"`
	FinishedAt    time.Time `json:"finished_at"`
	DurationMs    int64     `json:"duration_ms"`
}

type MacroResult struct {
	AffectedCount int
	Message       string
	PreviewPaths  []string
}
