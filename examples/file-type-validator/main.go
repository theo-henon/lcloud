package main

import (
	"encoding/json"

	plugincore "github.com/theo-henon/lcloud/internal/plugin"
	"github.com/theo-henon/lcloud/internal/volume"
	"github.com/theo-henon/lcloud/pkg/pluginsdk"
)

type validator struct{}

func main() {
	pluginsdk.Main(&validator{})
}

func (v *validator) HandleEvent(event plugincore.Event) (*plugincore.HandleResult, error) {
	if event.Type != pluginsdk.EventFileUploaded {
		return &plugincore.HandleResult{}, nil
	}

	var payload struct {
		Name    string         `json:"name"`
		Filters volume.Filters `json:"filters"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return &plugincore.HandleResult{
			Logs: []plugincore.LogRequest{{
				PluginID:  "file-type-validator",
				VolumeID:  event.VolumeID,
				EventType: event.Type,
				Level:     plugincore.LogLevelError,
				Message:   "unable to parse upload payload",
			}},
		}, nil
	}

	result, customEvent, level, message := validateUpload(payload.Name, payload.Filters)
	return &plugincore.HandleResult{
		Logs: []plugincore.LogRequest{{
			PluginID:  "file-type-validator",
			VolumeID:  event.VolumeID,
			EventType: customEvent,
			Level:     level,
			Message:   message,
		}},
		Emit: []plugincore.Event{
			plugincore.NewCustomEvent(customEvent, event.VolumeID, map[string]any{
				"filename": payload.Name,
				"result":   result,
			}),
		},
	}, nil
}

func validateUpload(filename string, filters volume.Filters) (result, eventType string, level plugincore.LogLevel, message string) {
	if err := volume.ValidateExtension(filters, filename); err != nil {
		return "failed", pluginsdk.EventValidationFailed, plugincore.LogLevelWarn, "validation failed for " + filename
	}
	return "passed", pluginsdk.EventValidationPassed, plugincore.LogLevelInfo, "validation passed for " + filename
}
