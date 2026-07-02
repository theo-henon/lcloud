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

func setupFileOpsTest(t *testing.T) (*Volume, *MetadataCache, *indexer.IndexManager, FileOpDeps) {
	t.Helper()
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

	meta := NewMetadataCache()
	indexManager := indexer.NewIndexManager()
	svc := &Service{db: db}
	deps := FileOpDeps{
		volumes: svc, paths: NewPathResolver(), metadata: meta,
		indexManager: indexManager, events: noopPublisher{},
	}
	return vol, meta, indexManager, deps
}

func TestMoveFile(t *testing.T) {
	vol, meta, indexManager, deps := setupFileOpsTest(t)
	userdata := filepath.Join(vol.RootPath, UserdataDir)
	require.NoError(t, os.WriteFile(filepath.Join(userdata, "photo.jpg"), []byte("data"), 0o644))

	record := FileMetadataRecord{
		ID: "file-1", Name: "photo.jpg", RelativePath: "photo.jpg",
		MimeType: "image/jpeg", SizeBytes: 4, ModifiedAt: time.Now().UTC(),
	}
	require.NoError(t, meta.Write(vol.RootPath, record))
	idx, err := indexManager.Get(vol.RootPath)
	require.NoError(t, err)
	require.NoError(t, idx.Index(indexer.FileMetadata{
		ID: record.ID, Name: record.Name, RelativePath: record.RelativePath,
		MimeType: record.MimeType, SizeBytes: record.SizeBytes, ModifiedAt: record.ModifiedAt,
	}))

	require.NoError(t, MoveFile(context.Background(), deps, vol, "photo.jpg", "images/photo.jpg"))
	_, err = os.Stat(filepath.Join(userdata, "images", "photo.jpg"))
	require.NoError(t, err)

	moved, err := meta.ReadByRelativePath(vol.RootPath, "images/photo.jpg")
	require.NoError(t, err)
	require.Equal(t, "images/photo.jpg", moved.RelativePath)
}

func TestRenameFile(t *testing.T) {
	vol, meta, indexManager, deps := setupFileOpsTest(t)
	userdata := filepath.Join(vol.RootPath, UserdataDir)
	require.NoError(t, os.WriteFile(filepath.Join(userdata, "notes.txt"), []byte("hello"), 0o644))

	record := FileMetadataRecord{
		ID: "file-2", Name: "notes.txt", RelativePath: "notes.txt",
		MimeType: "text/plain", SizeBytes: 5, ModifiedAt: time.Now().UTC(),
	}
	require.NoError(t, meta.Write(vol.RootPath, record))
	idx, err := indexManager.Get(vol.RootPath)
	require.NoError(t, err)
	require.NoError(t, idx.Index(indexer.FileMetadata{
		ID: record.ID, Name: record.Name, RelativePath: record.RelativePath,
		MimeType: record.MimeType, SizeBytes: record.SizeBytes, ModifiedAt: record.ModifiedAt,
	}))

	entry, err := RenameEntry(context.Background(), deps, vol, "notes.txt", "readme.txt")
	require.NoError(t, err)
	require.Equal(t, "readme.txt", entry.Name)
	require.Equal(t, "file", entry.Type)

	_, err = os.Stat(filepath.Join(userdata, "readme.txt"))
	require.NoError(t, err)

	renamed, err := meta.ReadByRelativePath(vol.RootPath, "readme.txt")
	require.NoError(t, err)
	require.Equal(t, "readme.txt", renamed.RelativePath)
}

func TestRenameDirectoryCascade(t *testing.T) {
	vol, meta, indexManager, deps := setupFileOpsTest(t)
	userdata := filepath.Join(vol.RootPath, UserdataDir)
	require.NoError(t, os.MkdirAll(filepath.Join(userdata, "photos", "2024"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(userdata, "photos", "2024", "img.jpg"), []byte("x"), 0o644))

	record := FileMetadataRecord{
		ID: "file-3", Name: "img.jpg", RelativePath: "photos/2024/img.jpg",
		MimeType: "image/jpeg", SizeBytes: 1, ModifiedAt: time.Now().UTC(),
	}
	require.NoError(t, meta.Write(vol.RootPath, record))
	idx, err := indexManager.Get(vol.RootPath)
	require.NoError(t, err)
	require.NoError(t, idx.Index(indexer.FileMetadata{
		ID: record.ID, Name: record.Name, RelativePath: record.RelativePath,
		MimeType: record.MimeType, SizeBytes: record.SizeBytes, ModifiedAt: record.ModifiedAt,
	}))

	entry, err := RenameEntry(context.Background(), deps, vol, "photos/2024", "2024-summer")
	require.NoError(t, err)
	require.Equal(t, "directory", entry.Type)
	require.Equal(t, "photos/2024-summer", entry.Path)

	_, err = os.Stat(filepath.Join(userdata, "photos", "2024-summer", "img.jpg"))
	require.NoError(t, err)

	updated, err := meta.ReadByRelativePath(vol.RootPath, "photos/2024-summer/img.jpg")
	require.NoError(t, err)
	require.Equal(t, "photos/2024-summer/img.jpg", updated.RelativePath)
}

func TestValidateEntryName(t *testing.T) {
	require.Error(t, validateEntryName(""))
	require.Error(t, validateEntryName("../bad"))
	require.Error(t, validateEntryName(`bad/name`))
	require.NoError(t, validateEntryName("valid-name.txt"))
}
