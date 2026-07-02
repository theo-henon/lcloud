package plugin

import (
	"context"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type VolumeProvider interface {
	GetVolumeConfig(ctx context.Context, volumeID uuid.UUID) (*VolumeConfigView, error)
	VolumeRootPath(ctx context.Context, volumeID uuid.UUID) (string, error)
}

type HostAPIServer struct {
	pluginID string
	bus      *EventBus
	registry *Registry
	volumes  VolumeProvider
}

func NewHostAPIServer(pluginID string, bus *EventBus, registry *Registry, volumes VolumeProvider) *HostAPIServer {
	return &HostAPIServer{
		pluginID: pluginID,
		bus:      bus,
		registry: registry,
		volumes:  volumes,
	}
}

func (h *HostAPIServer) Emit(ctx context.Context, event Event) error {
	h.bus.Publish(ctx, event)
	return nil
}

func (h *HostAPIServer) Log(ctx context.Context, req LogRequest) error {
	if req.PluginID == "" {
		req.PluginID = h.pluginID
	}
	_, err := h.registry.AppendLog(ctx, req)
	return err
}

func (h *HostAPIServer) GetVolumeConfig(ctx context.Context, volumeID string) (*VolumeConfigView, error) {
	id, err := uuid.Parse(volumeID)
	if err != nil {
		return nil, err
	}
	return h.volumes.GetVolumeConfig(ctx, id)
}

func (h *HostAPIServer) PluginDataDir(ctx context.Context, volumeID, pluginID string) (string, error) {
	if pluginID == "" {
		pluginID = h.pluginID
	}
	volID, err := uuid.Parse(volumeID)
	if err != nil {
		return "", err
	}
	root, err := h.volumes.VolumeRootPath(ctx, volID)
	if err != nil {
		return "", err
	}
	dir := filepath.Join(root, "plugins", pluginID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}
