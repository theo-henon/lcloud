package volume

import (
	"context"
	"time"

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

type FileTrashedEvent struct {
	VolumeID     uuid.UUID
	FileID       string
	OriginalPath string
	SizeBytes    int64
	DeletedAt    time.Time
}

type FileRestoredEvent struct {
	VolumeID     uuid.UUID
	FileID       string
	RestoredPath string
	SizeBytes    int64
}

type FileMovedEvent struct {
	VolumeID  uuid.UUID
	FromPath  string
	ToPath    string
	SizeBytes int64
}

type FileRenamedEvent struct {
	VolumeID  uuid.UUID
	OldPath   string
	NewPath   string
	EntryType string // "file" | "directory"
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
	FileTrashed(ctx context.Context, event FileTrashedEvent)
	FileRestored(ctx context.Context, event FileRestoredEvent)
	FileMoved(ctx context.Context, event FileMovedEvent)
	FileRenamed(ctx context.Context, event FileRenamedEvent)
	VolumeCreated(ctx context.Context, event VolumeCreatedEvent)
	VolumeUpdated(ctx context.Context, event VolumeUpdatedEvent)
	VolumeDeleted(ctx context.Context, event VolumeDeletedEvent)
}
