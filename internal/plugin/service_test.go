package plugin

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/config"
	"github.com/theo-henon/lcloud/internal/indexer"
	"github.com/theo-henon/lcloud/internal/volume"
)

func TestServiceStartStop(t *testing.T) {
	db := openTestDB(t)
	root := t.TempDir()
	cfg := &config.Config{StorageBasePath: root, StorageDiskPaths: []string{root}}
	disks := volume.NewDiskRegistry(cfg)
	volService := volume.NewService(db, disks, indexer.NewIndexManager())
	service := NewService(db, t.TempDir(), volService, nil)

	require.NoError(t, service.Start(context.Background()))
	service.Stop()
}

func TestServiceGetAndListLogs(t *testing.T) {
	db := openTestDB(t)
	service := NewService(db, t.TempDir(), nil, nil)

	_, err := service.Get("missing")
	require.ErrorIs(t, err, ErrPluginNotFound)

	claims := &auth.Claims{Role: auth.RoleAdmin, UserID: uuid.New()}
	logs, err := service.ListLogs(context.Background(), claims, LogFilter{Limit: 10})
	require.NoError(t, err)
	require.Empty(t, logs)
}

func TestVolumeProviderGetConfig(t *testing.T) {
	db := openTestDB(t)
	root := t.TempDir()
	cfg := &config.Config{StorageBasePath: root, StorageDiskPaths: []string{root}}
	disks := volume.NewDiskRegistry(cfg)
	volService := volume.NewService(db, disks, indexer.NewIndexManager())

	vol := &volume.Volume{
		ID:         uuid.New(),
		Name:       "Photos",
		RootPath:   root,
		QuotaBytes: 1000,
		Filters:    volume.Filters{Mode: volume.FilterModeAllow, Extensions: []string{".png"}},
	}
	require.NoError(t, db.Create(vol).Error)

	provider := newVolumeProvider(volService)
	view, err := provider.GetVolumeConfig(context.Background(), vol.ID)
	require.NoError(t, err)
	require.Equal(t, "Photos", view.Name)

	path, err := provider.VolumeRootPath(context.Background(), vol.ID)
	require.NoError(t, err)
	require.Equal(t, root, path)
}

func TestRuntimeDispatchError(t *testing.T) {
	db := openTestDB(t)
	registry := NewRegistry(db, t.TempDir(), nil)
	bus := NewEventBus()
	runtime := NewRuntime(registry, bus, stubVolumeProvider{}, nil)

	now := time.Now().UTC()
	record := Plugin{
		ID:            "demo",
		Name:          "Demo",
		Version:       "1.0.0",
		BinaryPath:    "/tmp/demo",
		ManifestPath:  "/tmp/demo.json",
		Enabled:       true,
		Status:        PluginStatusRunning,
		Subscriptions: StringArray{"file.uploaded"},
		DiscoveredAt:  now,
		UpdatedAt:     now,
	}
	require.NoError(t, db.Create(&record).Error)

	runtime.mu.Lock()
	runtime.active["demo"] = &runningPlugin{
		record: record,
		api:    mockPluginAPI{err: ErrPluginStartFailed},
	}
	runtime.mu.Unlock()

	runtime.dispatch(context.Background(), runtime.active["demo"], Event{Type: EventFileUploaded, VolumeID: uuid.NewString()})

	updated, err := registry.Get("demo")
	require.NoError(t, err)
	require.Equal(t, PluginStatusError, updated.Status)
}

func TestServiceSetEnabledDisable(t *testing.T) {
	db := openTestDB(t)
	now := time.Now().UTC()
	require.NoError(t, db.Create(&Plugin{
		ID: "demo", Name: "Demo", Version: "1", BinaryPath: "/tmp/demo", ManifestPath: "/tmp/demo.json",
		Enabled: true, Status: PluginStatusStopped, Subscriptions: StringArray{},
		DiscoveredAt: now, UpdatedAt: now,
	}).Error)

	service := NewService(db, t.TempDir(), nil, nil)
	require.NoError(t, service.SetEnabled(context.Background(), "demo", false))

	record, err := service.Get("demo")
	require.NoError(t, err)
	require.False(t, record.Enabled)
}

func TestRuntimeStopAll(t *testing.T) {
	db := openTestDB(t)
	registry := NewRegistry(db, t.TempDir(), nil)
	runtime := NewRuntime(registry, NewEventBus(), stubVolumeProvider{}, nil)
	runtime.StopAll()
}

func TestRuntimeDispatchSuccess(t *testing.T) {
	db := openTestDB(t)
	registry := NewRegistry(db, t.TempDir(), nil)
	runtime := NewRuntime(registry, NewEventBus(), stubVolumeProvider{}, nil)

	now := time.Now().UTC()
	record := Plugin{
		ID: "demo", Name: "Demo", Version: "1", BinaryPath: "/tmp/demo", ManifestPath: "/tmp/demo.json",
		Enabled: true, Status: PluginStatusRunning, Subscriptions: StringArray{"file.uploaded"},
		DiscoveredAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&record).Error)

	runtime.mu.Lock()
	runtime.active["demo"] = &runningPlugin{
		record: record,
		api: mockPluginAPI{
			result: &HandleResult{
				Logs: []LogRequest{{Message: "ok", Level: LogLevelInfo, PluginID: "demo"}},
				Emit: []Event{NewCustomEvent("validation.passed", "", nil)},
			},
		},
	}
	runtime.mu.Unlock()

	runtime.dispatch(context.Background(), runtime.active["demo"], Event{Type: EventFileUploaded})
}

type mockPluginAPI struct {
	err    error
	result *HandleResult
}

func (m mockPluginAPI) HandleEvent(_ context.Context, _ Event) (*HandleResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.result != nil {
		return m.result, nil
	}
	return &HandleResult{}, nil
}
