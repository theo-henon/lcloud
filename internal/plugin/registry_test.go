package plugin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"github.com/theo-henon/lcloud/internal/volume"
	"gorm.io/gorm"
)

func TestRegistryScanAndLoad(t *testing.T) {
	dir := t.TempDir()
	binaryPath := filepath.Join(dir, "demo-plugin")
	require.NoError(t, os.WriteFile(binaryPath, []byte("#!/bin/sh\n"), 0o755))

	manifest := Manifest{
		ID:          "demo-plugin",
		Name:        "Demo Plugin",
		Version:     "1.0.0",
		Description: "test",
		Subscribe:   []string{"file.uploaded"},
		Emit:        []string{"demo.event"},
	}
	raw, err := json.Marshal(manifest)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(binaryPath+".json", raw, 0o644))

	db := openTestDB(t)
	registry := NewRegistry(db, dir, nil)

	loaded, err := registry.ScanAndLoad()
	require.NoError(t, err)
	require.Len(t, loaded, 1)
	require.Equal(t, "demo-plugin", loaded[0].ID)
}

func TestRegistrySkipsMissingManifest(t *testing.T) {
	dir := t.TempDir()
	binaryPath := filepath.Join(dir, "orphan")
	require.NoError(t, os.WriteFile(binaryPath, []byte("#!/bin/sh\n"), 0o755))

	db := openTestDB(t)
	registry := NewRegistry(db, dir, nil)

	loaded, err := registry.ScanAndLoad()
	require.NoError(t, err)
	require.Empty(t, loaded)
}

func TestValidateManifest(t *testing.T) {
	require.Error(t, validateManifest(&Manifest{ID: "Bad ID", Name: "x", Version: "1"}))
	require.NoError(t, validateManifest(&Manifest{ID: "ok-plugin", Name: "x", Version: "1"}))
}

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "file:" + t.Name() + "?mode=memory&cache=private"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&Plugin{}, &PluginLogEntry{}, &volume.Volume{}))
	return db
}
