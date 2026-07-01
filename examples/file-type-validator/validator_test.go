package main

import (
	"testing"

	plugincore "github.com/theo-henon/lcloud/internal/plugin"
	"github.com/theo-henon/lcloud/internal/volume"
	"github.com/theo-henon/lcloud/pkg/pluginsdk"
)

func TestValidateUploadAllowed(t *testing.T) {
	result, eventType, level, message := validateUpload("photo.png", volume.Filters{
		Mode:       volume.FilterModeAllow,
		Extensions: []string{".png"},
	})
	if result != "passed" || eventType != pluginsdk.EventValidationPassed || level != plugincore.LogLevelInfo {
		t.Fatalf("unexpected validation result: %s %s %s %s", result, eventType, level, message)
	}
}

func TestValidateUploadRejected(t *testing.T) {
	result, eventType, level, _ := validateUpload("video.mp4", volume.Filters{
		Mode:       volume.FilterModeAllow,
		Extensions: []string{".png"},
	})
	if result != "failed" || eventType != pluginsdk.EventValidationFailed || level != plugincore.LogLevelWarn {
		t.Fatalf("expected failed validation")
	}
}
