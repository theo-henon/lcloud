package volume

import (
	"bytes"
	"image"
	"image/png"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/config"
	"github.com/theo-henon/lcloud/internal/indexer"
	"gorm.io/gorm"
)

func setupVolumeTest(t *testing.T) (*Service, *FileService, *auth.Claims, string) {
	t.Helper()

	storageRoot := t.TempDir()
	cfg := &config.Config{
		StorageBasePath:  storageRoot,
		StorageDiskPaths: []string{storageRoot},
		MaxUploadBytes:   10 * 1024 * 1024,
	}

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&Volume{}))

	indexManager := indexer.NewIndexManager()
	service := NewService(db, NewDiskRegistry(cfg), indexManager)
	files := NewFileService(service, indexManager)
	claims := &auth.Claims{UserID: uuid.New(), Role: auth.RoleUser}

	return service, files, claims, storageRoot
}

func TestVolumeCreateUploadDownloadFlow(t *testing.T) {
	service, files, claims, _ := setupVolumeTest(t)

	vol, err := service.Create(claims, CreateVolumeInput{
		Name:       "Photos",
		DiskPath:   service.disks.ResolvedPaths()[0],
		QuotaBytes: 1024 * 1024,
		Filters: Filters{
			Mode:       FilterModeAllow,
			Extensions: []string{".png"},
		},
	})
	require.NoError(t, err)

	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	require.NoError(t, png.Encode(&buf, img))

	entry, err := files.Upload(claims, vol.ID, ".", "photo.png", bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)
	require.Equal(t, "photo.png", entry.Name)

	listing, err := files.List(claims, vol.ID, ".")
	require.NoError(t, err)
	require.Len(t, listing.Entries, 1)

	_, err = files.Upload(claims, vol.ID, ".", "video.mp4", bytes.NewReader([]byte("fake")), 4)
	require.ErrorIs(t, err, ErrFileFilterRejected)

	file, _, err := files.OpenContent(claims, vol.ID, "photo.png")
	require.NoError(t, err)
	defer file.Close()
}

func TestVolumeDeleteRules(t *testing.T) {
	service, files, claims, _ := setupVolumeTest(t)
	adminClaims := &auth.Claims{UserID: claims.UserID, Role: auth.RoleAdmin}

	vol, err := service.Create(claims, CreateVolumeInput{
		Name:     "Temp",
		DiskPath: service.disks.ResolvedPaths()[0],
	})
	require.NoError(t, err)

	_, err = files.Upload(claims, vol.ID, ".", "note.txt", bytes.NewReader([]byte("hello")), 5)
	require.NoError(t, err)

	require.ErrorIs(t, service.Delete(claims, vol.ID, false), ErrVolumeNotEmpty)
	require.NoError(t, service.Delete(adminClaims, vol.ID, true))
}
