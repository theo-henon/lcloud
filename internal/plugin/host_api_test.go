package plugin

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/theo-henon/lcloud/internal/volume"
)

type stubVolumeProvider struct {
	vol *volume.Volume
}

func (s stubVolumeProvider) GetVolumeConfig(_ context.Context, volumeID uuid.UUID) (*VolumeConfigView, error) {
	if s.vol == nil || s.vol.ID != volumeID {
		return nil, volume.ErrVolumeNotFound
	}
	return &VolumeConfigView{
		ID:         s.vol.ID.String(),
		Name:       s.vol.Name,
		QuotaBytes: s.vol.QuotaBytes,
		Filters:    s.vol.Filters,
	}, nil
}

func (s stubVolumeProvider) VolumeRootPath(_ context.Context, volumeID uuid.UUID) (string, error) {
	if s.vol == nil || s.vol.ID != volumeID {
		return "", volume.ErrVolumeNotFound
	}
	return s.vol.RootPath, nil
}

func TestHostAPIPluginDataDir(t *testing.T) {
	dir := t.TempDir()
	volID := uuid.New()
	provider := stubVolumeProvider{
		vol: &volume.Volume{
			ID:       volID,
			Name:     "Photos",
			RootPath: dir,
		},
	}

	db := openTestDB(t)
	registry := NewRegistry(db, t.TempDir(), nil)
	bus := NewEventBus()
	host := NewHostAPIServer("demo", bus, registry, provider)

	path, err := host.PluginDataDir(context.Background(), volID.String(), "demo")
	require.NoError(t, err)
	require.Contains(t, path, "plugins/demo")
}

func TestHostAPILogRetention(t *testing.T) {
	db := openTestDB(t)
	registry := NewRegistry(db, t.TempDir(), nil)
	bus := NewEventBus()
	host := NewHostAPIServer("demo", bus, registry, stubVolumeProvider{})

	for i := 0; i < maxLogEntriesPerPlugin+5; i++ {
		require.NoError(t, host.Log(context.Background(), LogRequest{
			PluginID: "demo",
			Level:    LogLevelInfo,
			Message:  "entry",
		}))
	}

	var count int64
	require.NoError(t, db.Model(&PluginLogEntry{}).Where("plugin_id = ?", "demo").Count(&count).Error)
	require.Equal(t, int64(maxLogEntriesPerPlugin), count)
}
