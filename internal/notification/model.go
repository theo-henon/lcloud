package notification

import (
	"time"

	"github.com/google/uuid"
)

const (
	TypeVolumeUsageAlert      = "volume.usage_alert"
	TypeTaskFailed            = "task.failed"
	TypeVolumeDeletionRequest = "volume.deletion_requested"

	maxNotificationsPerUser = 200
)

type Notification struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null;index:idx_notif_user_dismissed_created,priority:1"`
	Type          string     `gorm:"type:varchar(64);not null"`
	Title         string     `gorm:"type:varchar(200);not null"`
	Body          string     `gorm:"type:text;not null"`
	LinkPath      string     `gorm:"type:varchar(500)"`
	VolumeID      *uuid.UUID `gorm:"type:uuid"`
	TaskID        *uuid.UUID `gorm:"type:uuid"`
	SourceEventID *string    `gorm:"type:varchar(64);index"`
	ReadAt        *time.Time `gorm:"index:idx_notif_user_read"`
	DismissedAt   *time.Time `gorm:"index:idx_notif_user_dismissed_created,priority:2"`
	CreatedAt     time.Time  `gorm:"index:idx_notif_user_dismissed_created,priority:3,sort:desc"`
}

func (Notification) TableName() string {
	return "notifications"
}

type Response struct {
	ID        uuid.UUID  `json:"id"`
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	LinkPath  string     `json:"link_path"`
	VolumeID  *uuid.UUID `json:"volume_id"`
	TaskID    *uuid.UUID `json:"task_id"`
	ReadAt    *time.Time `json:"read_at"`
	CreatedAt time.Time  `json:"created_at"`
}

func ToResponse(n Notification) Response {
	return Response{
		ID:        n.ID,
		Type:      n.Type,
		Title:     n.Title,
		Body:      n.Body,
		LinkPath:  n.LinkPath,
		VolumeID:  n.VolumeID,
		TaskID:    n.TaskID,
		ReadAt:    n.ReadAt,
		CreatedAt: n.CreatedAt,
	}
}
