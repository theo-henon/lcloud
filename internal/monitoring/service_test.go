package monitoring

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/config"
	"github.com/theo-henon/lcloud/internal/indexer"
	"github.com/theo-henon/lcloud/internal/volume"
	"gorm.io/gorm"
)

func setupMonitoringTest(t *testing.T) (*Service, *volume.Service, *volume.FileService, *auth.Claims, string) {
	t.Helper()

	storageRoot := t.TempDir()
	cfg := &config.Config{
		StorageBasePath:  storageRoot,
		StorageDiskPaths: []string{storageRoot},
		MaxUploadBytes:   10 * 1024 * 1024,
	}

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&volume.Volume{}))

	indexManager := indexer.NewIndexManager()
	volumeService := volume.NewService(db, volume.NewDiskRegistry(cfg), indexManager)
	monitoringService := NewService(volumeService, volume.NewDiskRegistry(cfg))
	fileService := volume.NewFileService(volumeService, indexManager, cfg.MaxUploadBytes, monitoringService.StatsCache())
	claims := &auth.Claims{UserID: uuid.New(), Role: auth.RoleUser}

	return monitoringService, volumeService, fileService, claims, storageRoot
}

func TestService_ComputeStatsAndOverview(t *testing.T) {
	monitoringService, volumeService, files, claims, storageRoot := setupMonitoringTest(t)

	vol, err := volumeService.Create(claims, volume.CreateVolumeInput{
		Name:       "Photos",
		DiskPath:   storageRoot,
		QuotaBytes: 50 * 1024 * 1024 * 1024,
		Filters: volume.Filters{
			Mode:       volume.FilterModeAllow,
			Extensions: []string{".png"},
		},
	})
	require.NoError(t, err)

	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	require.NoError(t, png.Encode(&buf, img))
	_, err = files.Upload(claims, vol.ID, ".", "photo.png", bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)

	stats, err := monitoringService.ComputeStats(vol)
	require.NoError(t, err)
	require.Equal(t, 1, stats.FileCount)
	require.Equal(t, CategoryImages, stats.ByCategory[0].Category)

	overview, err := monitoringService.GetOverview(context.Background(), claims)
	require.NoError(t, err)
	require.Len(t, overview.Disks, 1)
	require.Len(t, overview.Volumes, 1)
	require.Equal(t, CategoryImages, overview.Volumes[0].TopCategory)
}

func TestService_StatsInvalidatedOnUpload(t *testing.T) {
	monitoringService, volumeService, files, claims, storageRoot := setupMonitoringTest(t)

	vol, err := volumeService.Create(claims, volume.CreateVolumeInput{
		Name:     "Photos",
		DiskPath: storageRoot,
		Filters: volume.Filters{
			Mode:       volume.FilterModeAllow,
			Extensions: []string{".png"},
		},
	})
	require.NoError(t, err)

	_, err = monitoringService.ComputeStats(vol)
	require.NoError(t, err)

	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	require.NoError(t, png.Encode(&buf, img))
	_, err = files.Upload(claims, vol.ID, ".", "photo.png", bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)

	_, err = monitoringService.StatsCache().Read(vol.RootPath)
	require.Error(t, err)
}
