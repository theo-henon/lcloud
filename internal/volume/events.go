package volume

import (
	"context"

	"github.com/google/uuid"
)

type FileUploadedEvent struct {
	VolumeID     uuid.UUID
	Name         string
	RelativePath string
	MimeType     string
	SizeBytes    int64
	Filters      Filters
}

type FileDeletedEvent struct {
	VolumeID     uuid.UUID
	RelativePath string
	SizeBytes    int64
}

type VolumeCreatedEvent struct {
	Volume *Volume
}

type VolumeUpdatedEvent struct {
	Volume *Volume
}

type VolumeDeletedEvent struct {
	VolumeID uuid.UUID
	Name     string
}

type EventPublisher interface {
	FileUploaded(ctx context.Context, event FileUploadedEvent)
	FileDeleted(ctx context.Context, event FileDeletedEvent)
	VolumeCreated(ctx context.Context, event VolumeCreatedEvent)
	VolumeUpdated(ctx context.Context, event VolumeUpdatedEvent)
	VolumeDeleted(ctx context.Context, event VolumeDeletedEvent)
}
