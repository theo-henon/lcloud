package main

import (
	plugincore "github.com/theo-henon/lcloud/internal/plugin"
	"github.com/theo-henon/lcloud/pkg/pluginsdk"
)

type echoPlugin struct{}

func (e *echoPlugin) HandleEvent(event plugincore.Event) (*plugincore.HandleResult, error) {
	return &plugincore.HandleResult{
		Logs: []plugincore.LogRequest{{
			PluginID:  "mock-echo",
			VolumeID:  event.VolumeID,
			EventType: event.Type,
			Level:     plugincore.LogLevelInfo,
			Message:   "handled " + event.Type,
		}},
	}, nil
}

func main() {
	pluginsdk.Main(&echoPlugin{})
}
