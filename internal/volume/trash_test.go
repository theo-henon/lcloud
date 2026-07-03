package volume

import (
	"context"
	"errors"
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

func setupTrashTestVolume(t *testing.T) (*Volume, *Service, *indexer.IndexManager, string) {
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

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&Volume{}))

	vol := &Volume{
		ID: volID, Name: "test", OwnerID: ownerID, RootPath: root,
		DiskPath: root, UsedBytes: 100,
	}
	require.NoError(t, db.Create(vol).Error)

	svc := &Service{db: db}
	indexManager := indexer.NewIndexManager()
	return vol, svc, indexManager, root
}

func trashDeps(svc *Service, indexManager *indexer.IndexManager) TrashOpDeps {
	return TrashOpDeps{
		FileOpDeps: FileOpDeps{
			volumes:      svc,
			paths:        NewPathResolver(),
			metadata:     NewMetadataCache(),
			indexManager: indexManager,
			events:       noopPublisher{},
		},
		thumbnails: NewThumbnailGenerator(),
	}
}

func TestTrashFileInternalSoftDelete(t *testing.T) {
	vol, svc, indexManager, root := setupTrashTestVolume(t)
	deps := trashDeps(svc, indexManager)

	userdata := filepath.Join(root, UserdataDir, "photo.jpg")
	require.NoError(t, os.WriteFile(userdata, []byte("data"), 0o644))

	meta := NewMetadataCache()
	record := FileMetadataRecord{
		ID: "file-1", Name: "photo.jpg", RelativePath: "photo.jpg",
		MimeType: "image/jpeg", SizeBytes: 4, ModifiedAt: time.Now().UTC(),
	}
	require.NoError(t, meta.Write(root, record))

	idx, err := indexManager.Get(root)
	require.NoError(t, err)
	require.NoError(t, idx.Index(indexer.FileMetadata{
		ID: record.ID, Name: record.Name, RelativePath: record.RelativePath,
		MimeType: record.MimeType, SizeBytes: record.SizeBytes, ModifiedAt: record.ModifiedAt,
	}))

	require.NoError(t, TrashFileInternal(context.Background(), deps, vol, "photo.jpg"))

	_, err = os.Stat(userdata)
	require.True(t, os.IsNotExist(err))

	trashBlob := filepath.Join(root, TrashDir, "file-1", "photo.jpg")
	_, err = os.Stat(trashBlob)
	require.NoError(t, err)

	updatedVol, err := svc.GetByID(vol.ID)
	require.NoError(t, err)
	require.Equal(t, int64(100), updatedVol.UsedBytes)

	_, err = meta.ReadByRelativePath(root, "photo.jpg")
	require.Error(t, err)

	results, err := idx.Search(indexer.SearchQuery{Term: "photo"})
	require.NoError(t, err)
	require.Empty(t, results)
}

func TestRestoreTrashItem(t *testing.T) {
	vol, svc, indexManager, root := setupTrashTestVolume(t)
	deps := trashDeps(svc, indexManager)

	userdata := filepath.Join(root, UserdataDir, "docs", "note.txt")
	require.NoError(t, os.MkdirAll(filepath.Dir(userdata), 0o755))
	require.NoError(t, os.WriteFile(userdata, []byte("hello"), 0o644))

	meta := NewMetadataCache()
	record := FileMetadataRecord{
		ID: "file-2", Name: "note.txt", RelativePath: "docs/note.txt",
		MimeType: "text/plain", SizeBytes: 5, ModifiedAt: time.Now().UTC(),
	}
	require.NoError(t, meta.Write(root, record))

	require.NoError(t, TrashFileInternal(context.Background(), deps, vol, "docs/note.txt"))

	entry, err := RestoreTrashItemInternal(context.Background(), deps, vol, "file-2")
	require.NoError(t, err)
	require.Equal(t, "docs/note.txt", entry.Path)

	_, err = os.Stat(userdata)
	require.NoError(t, err)

	restored, err := meta.ReadByRelativePath(root, "docs/note.txt")
	require.NoError(t, err)
	require.Equal(t, "file-2", restored.ID)
}

func TestTrashFileIDPathTraversalRejected(t *testing.T) {
	vol, svc, indexManager, _ := setupTrashTestVolume(t)
	deps := trashDeps(svc, indexManager)

	for _, maliciousID := range []string{"", ".", "..", "../secret", `foo\bar`, "foo/bar", "..%2F..%2F"} {
		_, err := RestoreTrashItemInternal(context.Background(), deps, vol, maliciousID)
		require.Error(t, err, "restore %q", maliciousID)
		require.True(t, errors.Is(err, ErrPathTraversal) || errors.Is(err, ErrTrashItemNotFound), "restore %q: %v", maliciousID, err)

		err = PurgeTrashItemInternal(context.Background(), deps, vol, maliciousID)
		require.Error(t, err, "purge %q", maliciousID)
		require.True(t, errors.Is(err, ErrPathTraversal) || errors.Is(err, ErrTrashItemNotFound), "purge %q: %v", maliciousID, err)
	}
}

func TestRestorePathOccupied(t *testing.T) {
	vol, svc, indexManager, root := setupTrashTestVolume(t)
	deps := trashDeps(svc, indexManager)

	userdata := filepath.Join(root, UserdataDir, "dup.txt")
	require.NoError(t, os.WriteFile(userdata, []byte("original"), 0o644))

	meta := NewMetadataCache()
	record := FileMetadataRecord{
		ID: "file-3", Name: "dup.txt", RelativePath: "dup.txt",
		MimeType: "text/plain", SizeBytes: 8, ModifiedAt: time.Now().UTC(),
	}
	require.NoError(t, meta.Write(root, record))
	require.NoError(t, TrashFileInternal(context.Background(), deps, vol, "dup.txt"))

	require.NoError(t, os.WriteFile(userdata, []byte("blocker"), 0o644))

	_, err := RestoreTrashItemInternal(context.Background(), deps, vol, "file-3")
	require.ErrorIs(t, err, ErrPathOccupied)
}

func TestPurgeTrashItemFreesQuota(t *testing.T) {
	vol, svc, indexManager, root := setupTrashTestVolume(t)
	deps := trashDeps(svc, indexManager)

	userdata := filepath.Join(root, UserdataDir, "big.bin")
	require.NoError(t, os.WriteFile(userdata, []byte("0123456789"), 0o644))

	meta := NewMetadataCache()
	record := FileMetadataRecord{
		ID: "file-4", Name: "big.bin", RelativePath: "big.bin",
		MimeType: "application/octet-stream", SizeBytes: 10, ModifiedAt: time.Now().UTC(),
	}
	require.NoError(t, meta.Write(root, record))
	require.NoError(t, TrashFileInternal(context.Background(), deps, vol, "big.bin"))

	require.NoError(t, PurgeTrashItemInternal(context.Background(), deps, vol, "file-4"))

	_, err := os.Stat(filepath.Join(root, TrashDir, "file-4"))
	require.True(t, os.IsNotExist(err))

	updatedVol, err := svc.GetByID(vol.ID)
	require.NoError(t, err)
	require.Equal(t, int64(90), updatedVol.UsedBytes)
}

func TestPurgeTrashOlderThan(t *testing.T) {
	vol, svc, indexManager, root := setupTrashTestVolume(t)
	deps := trashDeps(svc, indexManager)

	oldDeleted := time.Now().UTC().AddDate(0, 0, -40)
	recentDeleted := time.Now().UTC().AddDate(0, 0, -2)

	for _, item := range []struct {
		id, name string
		deleted  time.Time
	}{
		{"old-item", "old.txt", oldDeleted},
		{"new-item", "new.txt", recentDeleted},
	} {
		itemDir := trashItemDir(root, item.id)
		require.NoError(t, os.MkdirAll(itemDir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(itemDir, item.name), []byte("x"), 0o644))
		require.NoError(t, writeTrashMeta(itemDir, TrashMeta{
			ID: item.id, Name: item.name, OriginalPath: item.name,
			MimeType: "text/plain", SizeBytes: 1, DeletedAt: item.deleted,
			ModifiedAt: item.deleted,
		}))
	}

	count, _, err := PurgeTrashOlderThan(context.Background(), deps, vol, 30, false)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	_, err = os.Stat(trashItemDir(root, "old-item"))
	require.True(t, os.IsNotExist(err))
	_, err = os.Stat(trashItemDir(root, "new-item"))
	require.NoError(t, err)
}

func TestLazyTrashDirCreation(t *testing.T) {
	root := t.TempDir()
	volID := uuid.New()
	ownerID := uuid.New()
	// Old volume without .trash/ in layout (only userdata)
	require.NoError(t, os.MkdirAll(filepath.Join(root, UserdataDir), 0o755))
	require.NoError(t, WriteVolumeConfig(root, VolumeConfig{
		ID: volID, Name: "test", OwnerID: ownerID, DiskPath: root,
		CreatedAt: time.Now().UTC(),
	}))

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&Volume{}))

	vol := &Volume{
		ID: volID, Name: "test", OwnerID: ownerID, RootPath: root,
		DiskPath: root, UsedBytes: 0,
	}
	require.NoError(t, db.Create(vol).Error)

	svc := &Service{db: db}
	indexManager := indexer.NewIndexManager()
	deps := trashDeps(svc, indexManager)

	_, err = os.Stat(trashRoot(root))
	require.True(t, os.IsNotExist(err))

	userdata := filepath.Join(root, UserdataDir, "solo.txt")
	require.NoError(t, os.WriteFile(userdata, []byte("x"), 0o644))
	require.NoError(t, TrashFileInternal(context.Background(), deps, vol, "solo.txt"))

	_, err = os.Stat(trashRoot(root))
	require.NoError(t, err)
}
