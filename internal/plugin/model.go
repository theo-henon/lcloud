package plugin

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/theo-henon/lcloud/internal/volume"
)

const maxLogEntriesPerPlugin = 500

type StringArray []string

func (s StringArray) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	return json.Marshal(s)
}

func (s *StringArray) Scan(value any) error {
	if value == nil {
		*s = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return errors.New("unsupported StringArray scan type")
	}
}

type JSON map[string]any

func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	return json.Marshal(j)
}

func (j *JSON) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	default:
		return errors.New("unsupported JSON scan type")
	}
}

type PluginStatus string

const (
	PluginStatusRunning PluginStatus = "running"
	PluginStatusStopped PluginStatus = "stopped"
	PluginStatusError   PluginStatus = "error"
)

type Plugin struct {
	ID            string       `gorm:"primaryKey" json:"id"`
	Name          string       `gorm:"not null" json:"name"`
	Version       string       `gorm:"not null" json:"version"`
	Description   string       `json:"description"`
	Author        string       `json:"author"`
	BinaryPath    string       `gorm:"not null" json:"-"`
	ManifestPath  string       `gorm:"not null" json:"-"`
	Enabled       bool         `gorm:"not null;default:true" json:"enabled"`
	Status        PluginStatus `gorm:"not null;default:stopped" json:"status"`
	Subscriptions StringArray  `gorm:"type:jsonb;not null" json:"subscriptions"`
	DeclaredEmits StringArray  `gorm:"type:jsonb" json:"declared_emits"`
	LastError     string       `json:"last_error"`
	DiscoveredAt  time.Time    `json:"discovered_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

type PluginLogEntry struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	PluginID  string     `gorm:"not null;index" json:"plugin_id"`
	VolumeID  *uuid.UUID `gorm:"type:uuid;index" json:"volume_id,omitempty"`
	EventType string     `json:"event_type"`
	Level     string     `gorm:"not null" json:"level"`
	Message   string     `gorm:"not null" json:"message"`
	Payload   JSON       `gorm:"type:jsonb" json:"payload,omitempty"`
	CreatedAt time.Time  `gorm:"index" json:"created_at"`
}

type LogLevel string

const (
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

type LogRequest struct {
	PluginID  string
	VolumeID  string
	EventType string
	Level     LogLevel
	Message   string
	Payload   map[string]any
}

type VolumeConfigView struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	QuotaBytes int64          `json:"quota_bytes"`
	Filters    volume.Filters `json:"filters"`
}

type Manifest struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Subscribe   []string `json:"subscribe"`
	Emit        []string `json:"emit"`
}
