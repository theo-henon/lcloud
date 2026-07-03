package volume

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/theo-henon/lcloud/internal/indexer"
	"gorm.io/gorm"
)

type noopPublisher struct{}

func (noopPublisher) FileUploaded(context.Context, FileUploadedEvent)   {}
func (noopPublisher) FileDeleted(context.Context, FileDeletedEvent)     {}
func (noopPublisher) FileTrashed(context.Context, FileTrashedEvent)     {}
func (noopPublisher) FileRestored(context.Context, FileRestoredEvent)   {}
func (noopPublisher) FileMoved(context.Context, FileMovedEvent)         {}
func (noopPublisher) FileRenamed(context.Context, FileRenamedEvent)   {}
func (noopPublisher) VolumeCreated(context.Context, VolumeCreatedEvent) {}
func (noopPublisher) VolumeUpdated(context.Context, VolumeUpdatedEvent) {}
func (noopPublisher) VolumeDeleted(context.Context, VolumeDeletedEvent) {}

func TestMacroOpsMoveAndDelete(t *testing.T) {
	root := t.TempDir()
	volID := uuid.New()
	ownerID := uuid.New()
	require.NoError(t, createVolumeLayout(root))
	cfg := VolumeConfig{
		ID: volID, Name: "test", OwnerID: ownerID, DiskPath: root,
		CreatedAt: time.Now().UTC(),
	}
	require.NoError(t, WriteVolumeConfig(root, cfg))

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&Volume{}))

	vol := &Volume{
		ID: volID, Name: "test", OwnerID: ownerID, RootPath: root,
		DiskPath: root, UsedBytes: 100,
	}
	require.NoError(t, db.Create(vol).Error)

	userdata := filepath.Join(root, UserdataDir)
	require.NoError(t, os.WriteFile(filepath.Join(userdata, "photo.jpg"), []byte("data"), 0o644))

	meta := NewMetadataCache()
	record := FileMetadataRecord{
		ID: "file-1", Name: "photo.jpg", RelativePath: "photo.jpg",
		MimeType: "image/jpeg", SizeBytes: 4, ModifiedAt: time.Now().UTC(),
	}
	require.NoError(t, meta.Write(root, record))

	indexManager := indexer.NewIndexManager()
	idx, err := indexManager.Get(root)
	require.NoError(t, err)
	require.NoError(t, idx.Index(indexer.FileMetadata{
		ID: record.ID, Name: record.Name, RelativePath: record.RelativePath,
		MimeType: record.MimeType, SizeBytes: record.SizeBytes, ModifiedAt: record.ModifiedAt,
	}))

	svc := &Service{db: db}
	ops := NewMacroOps(svc, indexManager, nil, noopPublisher{})

	require.NoError(t, ops.MoveFileInternal(context.Background(), vol, "photo.jpg", "images/photo.jpg"))
	_, err = os.Stat(filepath.Join(userdata, "images", "photo.jpg"))
	require.NoError(t, err)

	moved, err := meta.ReadByRelativePath(root, "images/photo.jpg")
	require.NoError(t, err)
	require.Equal(t, "images/photo.jpg", moved.RelativePath)

	require.NoError(t, ops.DeleteFileInternal(context.Background(), vol, "images/photo.jpg"))
	_, err = os.Stat(filepath.Join(userdata, "images", "photo.jpg"))
	require.True(t, os.IsNotExist(err))
}
