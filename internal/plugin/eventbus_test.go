package plugin

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/theo-henon/lcloud/internal/volume"
)

func TestEventBusPublishAsyncDoesNotBlock(t *testing.T) {
	bus := NewEventBus()
	var started atomic.Bool
	done := make(chan struct{})

	bus.Subscribe(EventFileUploaded, func(ctx context.Context, event Event) {
		started.Store(true)
		<-done
	})

	finished := make(chan struct{})
	go func() {
		bus.Publish(context.Background(), Event{Type: EventFileUploaded})
		close(finished)
	}()

	select {
	case <-finished:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Publish blocked on slow handler")
	}

	require.Eventually(t, func() bool { return started.Load() }, time.Second, 10*time.Millisecond)
	close(done)
}

func TestEventBusDispatchesMatchingHandlers(t *testing.T) {
	bus := NewEventBus()
	var count int
	bus.Subscribe(EventFileUploaded, func(ctx context.Context, event Event) {
		count++
	})
	bus.Subscribe(EventFileDeleted, func(ctx context.Context, event Event) {
		count += 10
	})

	bus.Publish(context.Background(), Event{Type: EventFileUploaded})
	require.Eventually(t, func() bool { return count == 1 }, time.Second, 10*time.Millisecond)
}

func TestEventBridgeImplementsVolumePublisher(t *testing.T) {
	bus := NewEventBus()
	bridge := NewEventBridge(bus)
	var received Event
	bus.Subscribe(EventFileUploaded, func(ctx context.Context, event Event) {
		received = event
	})

	bridge.FileUploaded(context.Background(), volumeFileUploadedFixture())
	require.Eventually(t, func() bool { return received.Type == EventFileUploaded }, time.Second, 10*time.Millisecond)
}

func volumeFileUploadedFixture() volume.FileUploadedEvent {
	return volume.FileUploadedEvent{
		VolumeID:     uuidMustParse("660e8400-e29b-41d4-a716-446655440001"),
		Name:         "photo.png",
		RelativePath: "photo.png",
		MimeType:     "image/png",
		SizeBytes:    1024,
	}
}

// avoid importing uuid in test only - use events helper
func uuidMustParse(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		panic(err)
	}
	return id
}
