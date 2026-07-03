package notification

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/theo-henon/lcloud/internal/plugin"
)

func (s *Service) RegisterSubscribers(pluginService *plugin.Service) {
	pluginService.Subscribe(plugin.EventVolumeAlertUsage, s.handleVolumeAlertUsage)
	pluginService.Subscribe(plugin.EventTaskFailed, s.handleTaskFailed)
}

func (s *Service) handleVolumeAlertUsage(ctx context.Context, event plugin.Event) {
	var payload struct {
		VolumeID     string `json:"volume_id"`
		UsagePercent int    `json:"usage_percent"`
		QuotaBytes   int64  `json:"quota_bytes"`
		UsedBytes    int64  `json:"used_bytes"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return
	}

	volumeID, err := uuid.Parse(payload.VolumeID)
	if err != nil {
		return
	}

	vol, err := s.volumes.GetByID(volumeID)
	if err != nil {
		return
	}

	sourceID := event.ID
	title := fmt.Sprintf("%s quota alert", vol.Name)
	body := fmt.Sprintf(
		"Storage is at %d%% of quota (%s / %s).",
		payload.UsagePercent,
		formatBytes(payload.UsedBytes),
		formatQuota(payload.QuotaBytes),
	)
	linkPath := fmt.Sprintf("/volumes/%s", volumeID)

	_ = s.create(ctx, createInput{
		UserID:        vol.OwnerID,
		Type:          TypeVolumeUsageAlert,
		Title:         title,
		Body:          body,
		LinkPath:      linkPath,
		VolumeID:      &volumeID,
		SourceEventID: &sourceID,
	})
}

func (s *Service) handleTaskFailed(ctx context.Context, event plugin.Event) {
	var payload struct {
		TaskID string `json:"task_id"`
		Macro  string `json:"macro"`
		Error  string `json:"error"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return
	}

	taskID, err := uuid.Parse(payload.TaskID)
	if err != nil {
		return
	}

	t, err := s.tasks.GetByID(taskID)
	if err != nil {
		return
	}

	sourceID := event.ID
	title := fmt.Sprintf("Task failed: %s", t.Name)
	body := fmt.Sprintf("Macro `%s` failed: %s", payload.Macro, payload.Error)
	linkPath := fmt.Sprintf("/tasks?task=%s", taskID)

	_ = s.create(ctx, createInput{
		UserID:        t.OwnerID,
		Type:          TypeTaskFailed,
		Title:         title,
		Body:          body,
		LinkPath:      linkPath,
		TaskID:        &taskID,
		SourceEventID: &sourceID,
	})
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatQuota(quotaBytes int64) string {
	if quotaBytes <= 0 {
		return "unlimited"
	}
	return formatBytes(quotaBytes)
}
