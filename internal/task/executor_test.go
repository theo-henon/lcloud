package task

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/theo-henon/lcloud/internal/config"
	"github.com/theo-henon/lcloud/internal/indexer"
	"github.com/theo-henon/lcloud/internal/monitoring"
	"github.com/theo-henon/lcloud/internal/volume"
	"gorm.io/gorm"
)

type noopPublisher struct{}

func (noopPublisher) FileUploaded(context.Context, volume.FileUploadedEvent)   {}
func (noopPublisher) FileDeleted(context.Context, volume.FileDeletedEvent)     {}
func (noopPublisher) FileMoved(context.Context, volume.FileMovedEvent)         {}
func (noopPublisher) FileRenamed(context.Context, volume.FileRenamedEvent)     {}
func (noopPublisher) VolumeCreated(context.Context, volume.VolumeCreatedEvent) {}
func (noopPublisher) VolumeUpdated(context.Context, volume.VolumeUpdatedEvent) {}
func (noopPublisher) VolumeDeleted(context.Context, volume.VolumeDeletedEvent) {}

func TestExecutorDeleteOldFilesDryRun(t *testing.T) {
	root := t.TempDir()
	volID := uuid.New()
	ownerID := uuid.New()
	require.NoError(t, createTestVolume(root, volID, ownerID))

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&volume.Volume{}))

	vol := &volume.Volume{
		ID: volID, Name: "test", OwnerID: ownerID, RootPath: root,
		DiskPath: root, UsedBytes: 0,
	}
	require.NoError(t, db.Create(vol).Error)

	oldTime := time.Now().UTC().AddDate(0, 0, -120)
	oldFile := filepath.Join(root, volume.UserdataDir, "old.txt")
	require.NoError(t, os.WriteFile(oldFile, []byte("x"), 0o644))
	require.NoError(t, os.Chtimes(oldFile, oldTime, oldTime))

	meta := volume.NewMetadataCache()
	require.NoError(t, meta.Write(root, volume.FileMetadataRecord{
		ID: "f1", Name: "old.txt", RelativePath: "old.txt",
		MimeType: "text/plain", SizeBytes: 1, ModifiedAt: oldTime,
	}))

	cfg := &config.Config{StorageBasePath: root, StorageDiskPaths: []string{root}}
	diskRegistry := volume.NewDiskRegistry(cfg)
	indexManager := indexer.NewIndexManager()
	volSvc := volume.NewService(db, diskRegistry, indexManager)
	mon := monitoring.NewService(volSvc, diskRegistry, nil, nil)
	ops := volume.NewMacroOps(volSvc, indexManager, nil, noopPublisher{})
	exec := NewExecutor(ops, mon, volSvc, nil)

	task := &Task{
		Macro:      MacroDeleteOldFiles,
		Scope:      ScopeVolume,
		VolumeID:   &volID,
		Parameters: Parameters{"days": 90, "dry_run": true},
	}
	result, err := exec.Run(context.Background(), task, false)
	require.NoError(t, err)
	require.Equal(t, 1, result.AffectedCount)
	_, err = os.Stat(oldFile)
	require.NoError(t, err)
}

func createTestVolume(root string, volID, ownerID uuid.UUID) error {
	if err := os.MkdirAll(filepath.Join(root, volume.UserdataDir), 0o755); err != nil {
		return err
	}
	return volume.WriteVolumeConfig(root, volume.VolumeConfig{
		ID: volID, Name: "test", OwnerID: ownerID, DiskPath: root,
		CreatedAt: time.Now().UTC(),
	})
}
