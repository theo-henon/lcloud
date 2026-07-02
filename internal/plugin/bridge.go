package plugin

import (
	"context"

	"github.com/theo-henon/lcloud/internal/volume"
)

type EventBridge struct {
	bus *EventBus
}

func NewEventBridge(bus *EventBus) *EventBridge {
	return &EventBridge{bus: bus}
}

func (b *EventBridge) FileUploaded(ctx context.Context, event volume.FileUploadedEvent) {
	b.bus.Publish(ctx, FromFileUploaded(event))
}

func (b *EventBridge) FileDeleted(ctx context.Context, event volume.FileDeletedEvent) {
	b.bus.Publish(ctx, FromFileDeleted(event))
}

func (b *EventBridge) FileMoved(ctx context.Context, event volume.FileMovedEvent) {
	b.bus.Publish(ctx, FromFileMoved(event))
}

func (b *EventBridge) FileRenamed(ctx context.Context, event volume.FileRenamedEvent) {
	b.bus.Publish(ctx, FromFileRenamed(event))
}

func (b *EventBridge) VolumeCreated(ctx context.Context, event volume.VolumeCreatedEvent) {
	b.bus.Publish(ctx, FromVolumeCreated(event))
}

func (b *EventBridge) VolumeUpdated(ctx context.Context, event volume.VolumeUpdatedEvent) {
	b.bus.Publish(ctx, FromVolumeUpdated(event))
}

func (b *EventBridge) VolumeDeleted(ctx context.Context, event volume.VolumeDeletedEvent) {
	b.bus.Publish(ctx, FromVolumeDeleted(event))
}

var _ volume.EventPublisher = (*EventBridge)(nil)
