package plugin

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/theo-henon/lcloud/internal/volume"
)

const (
	EventFileUploaded   = "file.uploaded"
	EventFileDeleted    = "file.deleted"
	EventFileMoved      = "file.moved"
	EventFileRenamed    = "file.renamed"
	EventVolumeCreated  = "volume.created"
	EventVolumeUpdated  = "volume.updated"
	EventVolumeDeleted  = "volume.deleted"
	EventTaskExecuted   = "task.executed"
	EventTaskFailed     = "task.failed"
	EventPluginRegistered   = "plugin.registered"
	EventPluginUnregistered = "plugin.unregistered"
)

type Event struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	VolumeID  string          `json:"volume_id,omitempty"`
	Payload   json.RawMessage `json:"payload"`
}

type HandleResult struct {
	Emit []Event      `json:"emit,omitempty"`
	Logs []LogRequest `json:"logs,omitempty"`
}

func newEvent(eventType, volumeID string, payload any) Event {
	raw, _ := json.Marshal(payload)
	return Event{
		ID:        uuid.NewString(),
		Type:      eventType,
		Timestamp: time.Now().UTC(),
		VolumeID:  volumeID,
		Payload:   raw,
	}
}

func FromFileUploaded(e volume.FileUploadedEvent) Event {
	return newEvent(EventFileUploaded, e.VolumeID.String(), map[string]any{
		"volume_id":     e.VolumeID.String(),
		"name":          e.Name,
		"relative_path": e.RelativePath,
		"mime_type":     e.MimeType,
		"size_bytes":    e.SizeBytes,
		"filters":       e.Filters,
	})
}

func FromFileDeleted(e volume.FileDeletedEvent) Event {
	return newEvent(EventFileDeleted, e.VolumeID.String(), map[string]any{
		"volume_id":     e.VolumeID.String(),
		"relative_path": e.RelativePath,
		"size_bytes":    e.SizeBytes,
	})
}

func FromVolumeCreated(e volume.VolumeCreatedEvent) Event {
	vol := e.Volume
	return newEvent(EventVolumeCreated, vol.ID.String(), map[string]any{
		"volume_id":   vol.ID.String(),
		"name":        vol.Name,
		"owner_id":    vol.OwnerID.String(),
		"filters":     vol.Filters,
		"quota_bytes": vol.QuotaBytes,
	})
}

func FromVolumeUpdated(e volume.VolumeUpdatedEvent) Event {
	vol := e.Volume
	return newEvent(EventVolumeUpdated, vol.ID.String(), map[string]any{
		"volume_id":   vol.ID.String(),
		"name":        vol.Name,
		"filters":     vol.Filters,
		"quota_bytes": vol.QuotaBytes,
	})
}

func FromVolumeDeleted(e volume.VolumeDeletedEvent) Event {
	return newEvent(EventVolumeDeleted, e.VolumeID.String(), map[string]any{
		"volume_id": e.VolumeID.String(),
		"name":      e.Name,
	})
}

func FromPluginRegistered(pluginID string) Event {
	return newEvent(EventPluginRegistered, "", map[string]any{"plugin_id": pluginID})
}

func FromPluginUnregistered(pluginID string) Event {
	return newEvent(EventPluginUnregistered, "", map[string]any{"plugin_id": pluginID})
}

func NewCustomEvent(eventType, volumeID string, payload any) Event {
	return newEvent(eventType, volumeID, payload)
}
