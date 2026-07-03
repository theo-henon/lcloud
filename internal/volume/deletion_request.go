package volume

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type DeletionRequestNotifier interface {
	NotifyDeletionRequest(ctx context.Context, req VolumeDeletionRequest) error
}

type DeletionRequestStatus string

const (
	DeletionRequestPending   DeletionRequestStatus = "pending"
	DeletionRequestDismissed DeletionRequestStatus = "dismissed"
)

type VolumeDeletionRequest struct {
	ID          uuid.UUID             `gorm:"type:uuid;primaryKey" json:"id"`
	VolumeID    uuid.UUID             `gorm:"type:uuid;not null;index" json:"volume_id"`
	UserID      uuid.UUID             `gorm:"type:uuid;not null;index" json:"user_id"`
	Status      DeletionRequestStatus `gorm:"type:varchar(16);not null;default:pending" json:"status"`
	CreatedAt   time.Time             `json:"created_at"`
	DismissedAt *time.Time            `json:"dismissed_at,omitempty"`
}

func (VolumeDeletionRequest) TableName() string {
	return "volume_deletion_requests"
}

type DeletionRequestResponse struct {
	ID         uuid.UUID `json:"id"`
	VolumeID   uuid.UUID `json:"volume_id"`
	VolumeName string    `json:"volume_name"`
	UserID     uuid.UUID `json:"user_id"`
	UserEmail  string    `json:"user_email"`
	CreatedAt  time.Time `json:"created_at"`
}
