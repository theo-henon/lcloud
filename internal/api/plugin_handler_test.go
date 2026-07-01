package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/config"
	"github.com/theo-henon/lcloud/internal/indexer"
	"github.com/theo-henon/lcloud/internal/monitoring"
	"github.com/theo-henon/lcloud/internal/plugin"
	"github.com/theo-henon/lcloud/internal/settings"
	"github.com/theo-henon/lcloud/internal/volume"
	"gorm.io/gorm"
)

func setupPluginRouter(t *testing.T) (*gin.Engine, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	dsn := "file:" + t.Name() + "?mode=memory&cache=private"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&auth.User{}, &auth.RefreshToken{}, &volume.Volume{}, &settings.InstanceSettings{}, &plugin.Plugin{}, &plugin.PluginLogEntry{}))

	service := auth.NewService(db, "01234567890123456789012345678901", 24, 7)
	require.NoError(t, service.SeedAdmin("admin@example.com", "adminpass1"))

	settingsService := settings.NewService(db)
	require.NoError(t, settingsService.EnsureDefaults())

	storageRoot := t.TempDir()
	cfg := &config.Config{
		StorageBasePath:  storageRoot,
		StorageDiskPaths: []string{storageRoot},
		MaxUploadBytes:   10 * 1024 * 1024,
	}
	diskRegistry := volume.NewDiskRegistry(cfg)
	indexManager := indexer.NewIndexManager()
	volumeService := volume.NewService(db, diskRegistry, indexManager)
	monitoringService := monitoring.NewService(volumeService, diskRegistry, settingsService)
	fileService := volume.NewFileService(volumeService, indexManager, cfg.MaxUploadBytes, monitoringService.StatsCache())
	pluginService := plugin.NewService(db, t.TempDir(), volumeService, nil)

	router := NewRouter(RouterConfig{
		AuthService:       service,
		DiskRegistry:      diskRegistry,
		VolumeService:     volumeService,
		FileService:       fileService,
		MonitoringService: monitoringService,
		SettingsService:   settingsService,
		PluginService:     pluginService,
		IndexManager:      indexManager,
		MaxUploadBytes:    cfg.MaxUploadBytes,
		GinMode:           gin.TestMode,
	})

	login, err := service.Login("admin@example.com", "adminpass1")
	require.NoError(t, err)
	return router, login.AccessToken
}

func TestPluginList(t *testing.T) {
	router, token := setupPluginRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/plugins", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Plugins []plugin.Plugin `json:"plugins"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.NotNil(t, resp.Plugins)
}

func TestPluginLogsEmpty(t *testing.T) {
	router, token := setupPluginRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/plugins/logs", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestPluginPatchNotFound(t *testing.T) {
	router, token := setupPluginRouter(t)

	body := []byte(`{"enabled":false}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/plugins/missing-plugin", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
}
