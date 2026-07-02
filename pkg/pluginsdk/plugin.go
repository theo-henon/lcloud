package pluginsdk

import (
	"context"
	"os"

	"github.com/hashicorp/go-plugin"
	plugincore "github.com/theo-henon/lcloud/internal/plugin"
)

type Handler interface {
	HandleEvent(event plugincore.Event) (*plugincore.HandleResult, error)
}

type pluginHandler struct {
	handler Handler
}

func (p pluginHandler) HandleEvent(_ context.Context, event plugincore.Event) (*plugincore.HandleResult, error) {
	return p.handler.HandleEvent(event)
}

// Serve starts the plugin subprocess entrypoint.
func Serve(handler Handler) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugincore.Handshake,
		Plugins: map[string]plugin.Plugin{
			plugincore.PluginName: &plugincore.LcloudPluginPlugin{
				Impl: pluginHandler{handler: handler},
			},
		},
	})
}

func Main(handler Handler) {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		os.Exit(0)
	}
	Serve(handler)
}
