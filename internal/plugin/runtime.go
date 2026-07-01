package plugin

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

type runningPlugin struct {
	record Plugin
	client *plugin.Client
	api    LcloudPlugin
}

type Runtime struct {
	mu       sync.RWMutex
	registry *Registry
	bus      *EventBus
	volumes  VolumeProvider
	logger   *slog.Logger
	active   map[string]*runningPlugin
}

func NewRuntime(registry *Registry, bus *EventBus, volumes VolumeProvider, logger *slog.Logger) *Runtime {
	if logger == nil {
		logger = slog.Default()
	}
	return &Runtime{
		registry: registry,
		bus:      bus,
		volumes:  volumes,
		logger:   logger,
		active:   make(map[string]*runningPlugin),
	}
}

func (r *Runtime) StartEnabled(ctx context.Context) error {
	plugins, err := r.registry.List()
	if err != nil {
		return err
	}
	for _, record := range plugins {
		if !record.Enabled {
			continue
		}
		if err := r.Start(ctx, record.ID); err != nil {
			r.logger.Error("plugin start failed", "id", record.ID, "error", err)
		}
	}
	return nil
}

func (r *Runtime) Start(ctx context.Context, pluginID string) error {
	record, err := r.registry.Get(pluginID)
	if err != nil {
		return err
	}

	r.Stop(pluginID)

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: Handshake,
		Plugins: map[string]plugin.Plugin{
			PluginName: &LcloudPluginPlugin{},
		},
		Cmd:    exec.Command(record.BinaryPath),
		Logger: hclog.NewNullLogger(),
	})

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		_ = r.registry.SetStatus(pluginID, PluginStatusError, err.Error())
		return fmt.Errorf("%w: %v", ErrPluginStartFailed, err)
	}

	raw, err := rpcClient.Dispense(PluginName)
	if err != nil {
		client.Kill()
		_ = r.registry.SetStatus(pluginID, PluginStatusError, err.Error())
		return fmt.Errorf("%w: %v", ErrPluginStartFailed, err)
	}

	api, ok := raw.(LcloudPlugin)
	if !ok {
		client.Kill()
		_ = r.registry.SetStatus(pluginID, PluginStatusError, "invalid plugin type")
		return ErrPluginStartFailed
	}

	r.mu.Lock()
	r.active[pluginID] = &runningPlugin{record: *record, client: client, api: api}
	r.mu.Unlock()

	_ = r.registry.SetStatus(pluginID, PluginStatusRunning, "")
	r.bus.Publish(ctx, FromPluginRegistered(pluginID))
	return nil
}

func (r *Runtime) Stop(pluginID string) {
	r.mu.Lock()
	running, ok := r.active[pluginID]
	if ok {
		delete(r.active, pluginID)
	}
	r.mu.Unlock()

	if !ok {
		return
	}

	running.client.Kill()
	_ = r.registry.SetStatus(pluginID, PluginStatusStopped, "")
	r.bus.Publish(context.Background(), FromPluginUnregistered(pluginID))
}

func (r *Runtime) StopAll() {
	plugins, err := r.registry.List()
	if err != nil {
		return
	}
	for _, record := range plugins {
		r.Stop(record.ID)
	}
}

func (r *Runtime) Subscribe() {
	r.bus.Subscribe("*", r.handleEvent)
}

func (r *Runtime) handleEvent(ctx context.Context, event Event) {
	if event.Type == EventPluginRegistered || event.Type == EventPluginUnregistered {
		return
	}

	r.mu.RLock()
	active := make([]*runningPlugin, 0, len(r.active))
	for _, p := range r.active {
		active = append(active, p)
	}
	r.mu.RUnlock()

	for _, p := range active {
		if !subscribesTo(p.record.Subscriptions, event.Type) {
			continue
		}
		go r.dispatch(ctx, p, event)
	}
}

func subscribesTo(subs []string, eventType string) bool {
	for _, sub := range subs {
		if sub == eventType || sub == "*" {
			return true
		}
	}
	return false
}

func (r *Runtime) dispatch(ctx context.Context, p *runningPlugin, event Event) {
	host := NewHostAPIServer(p.record.ID, r.bus, r.registry, r.volumes)
	if event.VolumeID != "" {
		_, _ = host.PluginDataDir(ctx, event.VolumeID, p.record.ID)
	}

	result, err := p.api.HandleEvent(ctx, event)
	if err != nil {
		r.logger.Error("plugin handle event failed", "plugin_id", p.record.ID, "event", event.Type, "error", err)
		_ = r.registry.SetStatus(p.record.ID, PluginStatusError, err.Error())
		_, _ = r.registry.AppendLog(ctx, LogRequest{
			PluginID:  p.record.ID,
			VolumeID:  event.VolumeID,
			EventType: event.Type,
			Level:     LogLevelError,
			Message:   err.Error(),
		})
		return
	}
	if result == nil {
		return
	}

	for _, logReq := range result.Logs {
		if logReq.PluginID == "" {
			logReq.PluginID = p.record.ID
		}
		_ = host.Log(ctx, logReq)
	}
	for _, emitted := range result.Emit {
		_ = host.Emit(ctx, emitted)
	}
}
