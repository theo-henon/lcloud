package plugin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStringArrayValueScan(t *testing.T) {
	var arr StringArray
	require.NoError(t, arr.Scan([]byte(`["file.uploaded"]`)))
	require.Equal(t, StringArray{"file.uploaded"}, arr)

	value, err := StringArray{"a", "b"}.Value()
	require.NoError(t, err)
	require.NotNil(t, value)
}

func TestJSONValueScan(t *testing.T) {
	var payload JSON
	require.NoError(t, payload.Scan([]byte(`{"ok":true}`)))
	require.True(t, payload["ok"].(bool))

	value, err := JSON{"k": "v"}.Value()
	require.NoError(t, err)
	require.NotNil(t, value)

	var nilJSON JSON
	require.NoError(t, nilJSON.Scan(nil))
}

func TestValidateManifestMissingFields(t *testing.T) {
	require.Error(t, validateManifest(&Manifest{ID: "ok", Name: "", Version: "1"}))
	require.Error(t, validateManifest(&Manifest{ID: "ok", Name: "n", Version: ""}))
}

func TestRegistryScanUpdatesExisting(t *testing.T) {
	dir := t.TempDir()
	binaryPath := filepath.Join(dir, "demo-plugin")
	require.NoError(t, os.WriteFile(binaryPath, []byte("#!/bin/sh\n"), 0o755))
	manifest := Manifest{ID: "demo-plugin", Name: "Demo", Version: "1.0.0", Subscribe: []string{"file.uploaded"}}
	raw, err := json.Marshal(manifest)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(binaryPath+".json", raw, 0o644))

	db := openTestDB(t)
	registry := NewRegistry(db, dir, nil)
	_, err = registry.ScanAndLoad()
	require.NoError(t, err)

	manifest.Version = "2.0.0"
	raw, err = json.Marshal(manifest)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(binaryPath+".json", raw, 0o644))

	loaded, err := registry.ScanAndLoad()
	require.NoError(t, err)
	require.Len(t, loaded, 1)
	require.Equal(t, "2.0.0", loaded[0].Version)
}

func TestRegistryGetNotFound(t *testing.T) {
	db := openTestDB(t)
	registry := NewRegistry(db, t.TempDir(), nil)
	_, err := registry.Get("missing")
	require.ErrorIs(t, err, ErrPluginNotFound)
}

func TestRegistrySetEnabledNotFound(t *testing.T) {
	db := openTestDB(t)
	registry := NewRegistry(db, t.TempDir(), nil)
	require.ErrorIs(t, registry.SetEnabled("missing", true), ErrPluginNotFound)
}

func TestRuntimeStartNotFound(t *testing.T) {
	db := openTestDB(t)
	registry := NewRegistry(db, t.TempDir(), nil)
	runtime := NewRuntime(registry, NewEventBus(), stubVolumeProvider{}, nil)
	err := runtime.Start(t.Context(), "missing")
	require.ErrorIs(t, err, ErrPluginNotFound)
}

func TestRuntimeStartInvalidBinary(t *testing.T) {
	db := openTestDB(t)
	registry := NewRegistry(db, t.TempDir(), nil)
	now := time.Now().UTC()
	require.NoError(t, db.Create(&Plugin{
		ID: "bad", Name: "Bad", Version: "1", BinaryPath: "/no/such/binary", ManifestPath: "/tmp/x.json",
		Enabled: true, Status: PluginStatusStopped, Subscriptions: StringArray{"file.uploaded"},
		DiscoveredAt: now, UpdatedAt: now,
	}).Error)

	runtime := NewRuntime(registry, NewEventBus(), stubVolumeProvider{}, nil)
	err := runtime.Start(t.Context(), "bad")
	require.ErrorIs(t, err, ErrPluginStartFailed)
}
