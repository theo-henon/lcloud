package volume

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	VolumeJSONFile    = ".volume.json"
	UserdataDir       = "userdata"
	CacheDir          = "cache"
	IndexDir          = "index"
	ThumbnailsDir     = "thumbnails"
	MetadataDir       = "metadata"
	PluginsDir        = "plugins"
	LogsDir           = "logs"
)

type FilterMode string

const (
	FilterModeAllow FilterMode = "allow"
	FilterModeBlock FilterMode = "block"
)

type Filters struct {
	Mode       FilterMode `json:"mode"`
	Extensions []string   `json:"extensions"`
}

func (f Filters) Value() (driver.Value, error) {
	return json.Marshal(f)
}

func (f *Filters) Scan(value any) error {
	if value == nil {
		*f = Filters{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("filters: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, f)
}

type EncryptionConfig struct {
	Enabled bool   `json:"enabled"`
	Method  *string `json:"method"`
}

type VolumeConfig struct {
	ID         uuid.UUID        `json:"id"`
	Name       string           `json:"name"`
	OwnerID    uuid.UUID        `json:"owner_id"`
	QuotaBytes int64            `json:"quota_bytes"`
	Filters    Filters          `json:"filters"`
	Encryption EncryptionConfig `json:"encryption"`
	CreatedAt  time.Time        `json:"created_at"`
	DiskPath   string           `json:"disk_path"`
	UsedBytes  int64            `json:"used_bytes,omitempty"`
}

type Volume struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name       string    `gorm:"not null" json:"name"`
	OwnerID    uuid.UUID `gorm:"type:uuid;not null;index" json:"owner_id"`
	DiskPath   string    `gorm:"not null" json:"disk_path"`
	RootPath   string    `gorm:"not null" json:"root_path"`
	QuotaBytes int64     `gorm:"not null;default:0" json:"quota_bytes"`
	UsedBytes  int64     `gorm:"not null;default:0" json:"used_bytes"`
	Filters    Filters   `gorm:"type:jsonb;not null" json:"filters"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (Volume) TableName() string {
	return "volumes"
}

func (v *Volume) ToConfig() VolumeConfig {
	return VolumeConfig{
		ID:         v.ID,
		Name:       v.Name,
		OwnerID:    v.OwnerID,
		QuotaBytes: v.QuotaBytes,
		Filters:    v.Filters,
		Encryption: EncryptionConfig{Enabled: false, Method: nil},
		CreatedAt:  v.CreatedAt,
		DiskPath:   v.DiskPath,
		UsedBytes:  v.UsedBytes,
	}
}

func volumeFromConfig(cfg VolumeConfig, rootPath string) *Volume {
	return &Volume{
		ID:         cfg.ID,
		Name:       cfg.Name,
		OwnerID:    cfg.OwnerID,
		DiskPath:   cfg.DiskPath,
		RootPath:   rootPath,
		QuotaBytes: cfg.QuotaBytes,
		UsedBytes:  cfg.UsedBytes,
		Filters:    cfg.Filters,
		CreatedAt:  cfg.CreatedAt,
		UpdatedAt:  time.Now().UTC(),
	}
}
