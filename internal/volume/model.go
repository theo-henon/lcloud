package volume

import (
	"time"

	"github.com/google/uuid"
)

type Volume struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name       string    `gorm:"not null" json:"name"`
	OwnerID    uuid.UUID `gorm:"type:uuid;not null;index" json:"owner_id"`
	DiskPath   string    `gorm:"not null" json:"disk_path"`
	QuotaBytes int64     `gorm:"not null;default:0" json:"quota_bytes"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (Volume) TableName() string {
	return "volumes"
}
