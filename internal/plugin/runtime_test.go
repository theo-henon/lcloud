package plugin

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/theo-henon/lcloud/internal/volume"
)

func TestRuntimeStartAndDispatch(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("plugin subprocess test skipped on windows")
	}

	binary := buildMockPlugin(t)
	pluginsDir := t.TempDir()
	binaryDest := filepath.Join(pluginsDir, "mock-echo")
	require.NoError(t, os.WriteFile(binaryDest, binary, 0o755))

	manifest := Manifest{
		ID:        "mock-echo",
		Name:      "Mock Echo",
		Version:   "1.0.0",
		Subscribe: []string{"file.uploaded"},
	}
	raw, err := json.Marshal(manifest)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(binaryDest+".json", raw, 0o644))

	db := openTestDB(t)
	registry := NewRegistry(db, pluginsDir, nil)
	bus := NewEventBus()
	volID := uuid.New()
	root := t.TempDir()
	provider := stubVolumeProvider{vol: &volume.Volume{ID: volID, RootPath: root, Name: "Test"}}
	runtime := NewRuntime(registry, bus, provider, nil)
	runtime.Subscribe()

	_, err = registry.ScanAndLoad()
	require.NoError(t, err)
	require.NoError(t, runtime.Start(context.Background(), "mock-echo"))

	record, err := registry.Get("mock-echo")
	require.NoError(t, err)
	require.Equal(t, PluginStatusRunning, record.Status)

	bus.Publish(context.Background(), Event{
		Type:     EventFileUploaded,
		VolumeID: volID.String(),
	})

	require.Eventually(t, func() bool {
		entries, err := registry.ListLogs(context.Background(), LogFilter{PluginID: "mock-echo", Limit: 10, IsAdmin: true})
		return err == nil && len(entries) > 0
	}, 5*time.Second, 100*time.Millisecond)

	runtime.Stop("mock-echo")
	stopped, err := registry.Get("mock-echo")
	require.NoError(t, err)
	require.Equal(t, PluginStatusStopped, stopped.Status)
}

func buildMockPlugin(t *testing.T) []byte {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	pkgDir := filepath.Dir(file)
	out := filepath.Join(t.TempDir(), "mock-echo-bin")
	cmd := exec.Command("go", "build", "-o", out, "./testdata/mockplugin")
	cmd.Dir = pkgDir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
	data, err := os.ReadFile(out)
	require.NoError(t, err)
	return data
}

func TestSubscribesTo(t *testing.T) {
	require.True(t, subscribesTo([]string{"file.uploaded"}, EventFileUploaded))
	require.False(t, subscribesTo([]string{"file.deleted"}, EventFileUploaded))
	require.True(t, subscribesTo([]string{"*"}, EventFileUploaded))
}

func TestEventConstructors(t *testing.T) {
	volID := uuid.New()
	event := FromFileUploaded(volume.FileUploadedEvent{
		VolumeID: volID,
		Name:     "a.png",
	})
	require.Equal(t, EventFileUploaded, event.Type)
	require.Equal(t, volID.String(), event.VolumeID)

	custom := NewCustomEvent("validation.passed", volID.String(), map[string]any{"ok": true})
	require.Equal(t, "validation.passed", custom.Type)
}

func TestServicePublisher(t *testing.T) {
	db := openTestDB(t)
	service := NewService(db, t.TempDir(), nil, nil)

	var received bool
	service.bus.Subscribe(EventVolumeCreated, func(ctx context.Context, event Event) {
		received = true
	})

	service.Publisher().VolumeCreated(context.Background(), volume.VolumeCreatedEvent{
		Volume: &volume.Volume{ID: uuid.New(), Name: "demo"},
	})
	require.Eventually(t, func() bool { return received }, time.Second, 10*time.Millisecond)
}

func TestServiceListAndSetEnabledMissing(t *testing.T) {
	db := openTestDB(t)
	service := NewService(db, t.TempDir(), nil, nil)

	plugins, err := service.List()
	require.NoError(t, err)
	require.Empty(t, plugins)

	err = service.SetEnabled(context.Background(), "missing", false)
	require.ErrorIs(t, err, ErrPluginNotFound)
}

func TestBridgeAllEvents(t *testing.T) {
	bus := NewEventBus()
	bridge := NewEventBridge(bus)
	volID := uuid.New()
	vol := &volume.Volume{ID: volID, Name: "demo"}

	types := make(map[string]struct{})
	bus.Subscribe("*", func(ctx context.Context, event Event) {
		types[event.Type] = struct{}{}
	})

	bridge.FileUploaded(context.Background(), volume.FileUploadedEvent{VolumeID: volID, Name: "a"})
	bridge.FileDeleted(context.Background(), volume.FileDeletedEvent{VolumeID: volID})
	bridge.VolumeCreated(context.Background(), volume.VolumeCreatedEvent{Volume: vol})
	bridge.VolumeUpdated(context.Background(), volume.VolumeUpdatedEvent{Volume: vol})
	bridge.VolumeDeleted(context.Background(), volume.VolumeDeletedEvent{VolumeID: volID, Name: "demo"})

	require.Eventually(t, func() bool {
		return len(types) >= 5
	}, time.Second, 10*time.Millisecond)
}

func TestHostAPIEmit(t *testing.T) {
	db := openTestDB(t)
	registry := NewRegistry(db, t.TempDir(), nil)
	bus := NewEventBus()
	host := NewHostAPIServer("demo", bus, registry, stubVolumeProvider{})

	var emitted bool
	bus.Subscribe("custom.event", func(ctx context.Context, event Event) {
		emitted = true
	})

	require.NoError(t, host.Emit(context.Background(), NewCustomEvent("custom.event", "", nil)))
	require.Eventually(t, func() bool { return emitted }, time.Second, 10*time.Millisecond)
}

func TestRegistryListLogsOwnerScope(t *testing.T) {
	db := openTestDB(t)
	registry := NewRegistry(db, t.TempDir(), nil)
	ownerID := uuid.New()
	volID := uuid.New()
	require.NoError(t, db.Create(&volume.Volume{ID: volID, Name: "mine", OwnerID: ownerID, RootPath: t.TempDir()}).Error)

	_, err := registry.AppendLog(context.Background(), LogRequest{
		PluginID: "demo", VolumeID: volID.String(), Level: LogLevelInfo, Message: "scoped",
	})
	require.NoError(t, err)

	entries, err := registry.ListLogs(context.Background(), LogFilter{
		OwnerID: &ownerID, IsAdmin: false, Limit: 10,
	})
	require.NoError(t, err)
	require.Len(t, entries, 1)
}

func TestRegistrySetEnabledAndStatus(t *testing.T) {
	db := openTestDB(t)
	registry := NewRegistry(db, t.TempDir(), nil)

	record := Plugin{
		ID:            "demo",
		Name:          "Demo",
		Version:       "1.0.0",
		BinaryPath:    "/tmp/demo",
		ManifestPath:  "/tmp/demo.json",
		Enabled:       true,
		Status:        PluginStatusStopped,
		Subscriptions: StringArray{"file.uploaded"},
		DiscoveredAt:  time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	require.NoError(t, db.Create(&record).Error)

	require.NoError(t, registry.SetEnabled("demo", false))
	require.NoError(t, registry.SetStatus("demo", PluginStatusError, "boom"))

	updated, err := registry.Get("demo")
	require.NoError(t, err)
	require.False(t, updated.Enabled)
	require.Equal(t, PluginStatusError, updated.Status)
	require.Equal(t, "boom", updated.LastError)
}
