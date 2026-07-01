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

func setupTestRouter(t *testing.T) (*gin.Engine, *auth.Service, *volume.Service, string) {
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

	return router, service, volumeService, storageRoot
}

func setupTestRouterLegacy(t *testing.T) (*gin.Engine, *auth.Service) {
	router, service, _, _ := setupTestRouter(t)
	return router, service
}

func TestHealth(t *testing.T) {
	router, _ := setupTestRouterLegacy(t)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestMeUnauthorized(t *testing.T) {
	router, _ := setupTestRouterLegacy(t)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestLoginBadRequest(t *testing.T) {
	router, _ := setupTestRouterLegacy(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthFlow(t *testing.T) {
	router, service := setupTestRouterLegacy(t)

	loginBody := []byte(`{"email":"admin@example.com","password":"adminpass1"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var loginResp auth.LoginResult
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &loginResp))

	meReq := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+loginResp.AccessToken)
	meRec := httptest.NewRecorder()
	router.ServeHTTP(meRec, meReq)
	require.Equal(t, http.StatusOK, meRec.Code)

	refreshBody, err := json.Marshal(map[string]string{"refresh_token": loginResp.RefreshToken})
	require.NoError(t, err)
	refreshReq := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewReader(refreshBody))
	refreshReq.Header.Set("Content-Type", "application/json")
	refreshRec := httptest.NewRecorder()
	router.ServeHTTP(refreshRec, refreshReq)
	require.Equal(t, http.StatusOK, refreshRec.Code)

	var refreshResp auth.RefreshResult
	require.NoError(t, json.Unmarshal(refreshRec.Body.Bytes(), &refreshResp))

	logoutBody, err := json.Marshal(map[string]string{"refresh_token": refreshResp.RefreshToken})
	require.NoError(t, err)
	logoutReq := httptest.NewRequest(http.MethodPost, "/api/auth/logout", bytes.NewReader(logoutBody))
	logoutReq.Header.Set("Content-Type", "application/json")
	logoutRec := httptest.NewRecorder()
	router.ServeHTTP(logoutRec, logoutReq)
	require.Equal(t, http.StatusOK, logoutRec.Code)

	_, err = service.Refresh(refreshResp.RefreshToken)
	require.ErrorIs(t, err, auth.ErrRevokedToken)
}

func TestAdminCreateUserForbiddenForRegularUser(t *testing.T) {
	router, service := setupTestRouterLegacy(t)

	_, err := service.CreateUser("user@example.com", "password123", auth.RoleUser)
	require.NoError(t, err)

	login, err := service.Login("user@example.com", "password123")
	require.NoError(t, err)

	body := []byte(`{"email":"new@example.com","password":"password123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+login.AccessToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestAdminCreateUserInvalidRole(t *testing.T) {
	router, _ := setupTestRouterLegacy(t)

	loginBody := []byte(`{"email":"admin@example.com","password":"adminpass1"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)
	require.Equal(t, http.StatusOK, loginRec.Code)

	var loginResp auth.LoginResult
	require.NoError(t, json.Unmarshal(loginRec.Body.Bytes(), &loginResp))

	createBody := []byte(`{"email":"member@example.com","password":"password123","role":"superadmin"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/admin/users", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+loginResp.AccessToken)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusBadRequest, createRec.Code)
}

func TestAdminCreateUser(t *testing.T) {
	router, _ := setupTestRouterLegacy(t)

	loginBody := []byte(`{"email":"admin@example.com","password":"adminpass1"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)
	require.Equal(t, http.StatusOK, loginRec.Code)

	var loginResp auth.LoginResult
	require.NoError(t, json.Unmarshal(loginRec.Body.Bytes(), &loginResp))

	createBody := []byte(`{"email":"member@example.com","password":"password123","role":"user"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/admin/users", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+loginResp.AccessToken)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)
}

func TestAdminPatchSettings(t *testing.T) {
	router, _ := setupTestRouterLegacy(t)

	loginBody := []byte(`{"email":"admin@example.com","password":"adminpass1"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)
	require.Equal(t, http.StatusOK, loginRec.Code)

	var loginResp auth.LoginResult
	require.NoError(t, json.Unmarshal(loginRec.Body.Bytes(), &loginResp))

	patchBody := []byte(`{"mask_disk_names":true}`)
	patchReq := httptest.NewRequest(http.MethodPatch, "/api/admin/settings", bytes.NewReader(patchBody))
	patchReq.Header.Set("Content-Type", "application/json")
	patchReq.Header.Set("Authorization", "Bearer "+loginResp.AccessToken)
	patchRec := httptest.NewRecorder()
	router.ServeHTTP(patchRec, patchReq)
	require.Equal(t, http.StatusOK, patchRec.Code)

	var settingsResp struct {
		MaskDiskNames bool `json:"mask_disk_names"`
	}
	require.NoError(t, json.Unmarshal(patchRec.Body.Bytes(), &settingsResp))
	require.True(t, settingsResp.MaskDiskNames)

	getReq := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	getReq.Header.Set("Authorization", "Bearer "+loginResp.AccessToken)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)
	require.Equal(t, http.StatusOK, getRec.Code)
	require.NoError(t, json.Unmarshal(getRec.Body.Bytes(), &settingsResp))
	require.True(t, settingsResp.MaskDiskNames)
}
