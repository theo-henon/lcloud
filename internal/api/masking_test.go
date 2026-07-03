package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/monitoring"
	"github.com/theo-henon/lcloud/internal/volume"
)

func TestVolumeCreateAllowedForRegularUserWithDiskID(t *testing.T) {
	router, service, _, diskPath := setupTestRouter(t)

	_, err := service.CreateUser("user@example.com", "password123", auth.RoleUser)
	require.NoError(t, err)

	login, err := service.Login("user@example.com", "password123")
	require.NoError(t, err)

	body := []byte(`{"name":"Photos","disk_id":"storage-1","quota_bytes":0,"filters":{}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/volumes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+login.AccessToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var vol volume.Volume
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &vol))
	require.Equal(t, "Photos", vol.Name)
	require.Equal(t, diskPath, vol.DiskPath)
}

func TestVolumeDeleteForbiddenForRegularUser(t *testing.T) {
	router, service, volumeService, diskPath := setupTestRouter(t)

	user, err := service.CreateUser("user@example.com", "password123", auth.RoleUser)
	require.NoError(t, err)

	claims := &auth.Claims{UserID: user.ID, Role: auth.RoleUser}
	vol, err := volumeService.Create(claims, volume.CreateVolumeInput{
		Name:     "Photos",
		DiskPath: diskPath,
	})
	require.NoError(t, err)

	login, err := service.Login("user@example.com", "password123")
	require.NoError(t, err)

	delReq := httptest.NewRequest(http.MethodDelete, "/api/volumes/"+vol.ID.String(), nil)
	delReq.Header.Set("Authorization", "Bearer "+login.AccessToken)
	delRec := httptest.NewRecorder()
	router.ServeHTTP(delRec, delReq)
	require.Equal(t, http.StatusForbidden, delRec.Code)
}

func TestDiskMasking_HidesVolumePathsForRegularUser(t *testing.T) {
	router, service, volumeService, diskPath := setupTestRouter(t)

	adminLogin, err := service.Login("admin@example.com", "adminpass1")
	require.NoError(t, err)

	patchReq := httptest.NewRequest(
		http.MethodPatch,
		"/api/admin/settings",
		bytes.NewReader([]byte(`{"mask_disk_names":true}`)),
	)
	patchReq.Header.Set("Content-Type", "application/json")
	patchReq.Header.Set("Authorization", "Bearer "+adminLogin.AccessToken)
	patchRec := httptest.NewRecorder()
	router.ServeHTTP(patchRec, patchReq)
	require.Equal(t, http.StatusOK, patchRec.Code)

	member, err := service.CreateUser("member@example.com", "password123", auth.RoleUser)
	require.NoError(t, err)

	vol, err := volumeService.Create(
		&auth.Claims{UserID: member.ID, Role: auth.RoleUser},
		volume.CreateVolumeInput{
			Name:       "Photos",
			DiskPath:   diskPath,
			QuotaBytes: 50 * 1024 * 1024 * 1024,
			Filters:    volume.Filters{Mode: volume.FilterModeAllow, Extensions: []string{".png"}},
		},
	)
	require.NoError(t, err)

	userLogin, err := service.Login("member@example.com", "password123")
	require.NoError(t, err)

	getReq := httptest.NewRequest(http.MethodGet, "/api/volumes/"+vol.ID.String(), nil)
	getReq.Header.Set("Authorization", "Bearer "+userLogin.AccessToken)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)
	require.Equal(t, http.StatusOK, getRec.Code)

	var volumeResp struct {
		DiskPath string `json:"disk_path"`
		RootPath string `json:"root_path"`
	}
	require.NoError(t, json.Unmarshal(getRec.Body.Bytes(), &volumeResp))
	require.Empty(t, volumeResp.DiskPath)
	require.Empty(t, volumeResp.RootPath)

	listReq := httptest.NewRequest(http.MethodGet, "/api/volumes", nil)
	listReq.Header.Set("Authorization", "Bearer "+userLogin.AccessToken)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)
	require.Equal(t, http.StatusOK, listRec.Code)

	var listResp struct {
		Volumes []struct {
			DiskPath string `json:"disk_path"`
			RootPath string `json:"root_path"`
		} `json:"volumes"`
	}
	require.NoError(t, json.Unmarshal(listRec.Body.Bytes(), &listResp))
	require.Len(t, listResp.Volumes, 1)
	require.Empty(t, listResp.Volumes[0].DiskPath)
	require.Empty(t, listResp.Volumes[0].RootPath)

	userDisksReq := httptest.NewRequest(http.MethodGet, "/api/disks", nil)
	userDisksReq.Header.Set("Authorization", "Bearer "+userLogin.AccessToken)
	userDisksRec := httptest.NewRecorder()
	router.ServeHTTP(userDisksRec, userDisksReq)
	require.Equal(t, http.StatusOK, userDisksRec.Code)

	var disksResp struct {
		Disks []struct {
			Path string `json:"path"`
		} `json:"disks"`
	}
	require.NoError(t, json.Unmarshal(userDisksRec.Body.Bytes(), &disksResp))
	require.Empty(t, disksResp.Disks[0].Path)

	adminDisksReq := httptest.NewRequest(http.MethodGet, "/api/disks", nil)
	adminDisksReq.Header.Set("Authorization", "Bearer "+adminLogin.AccessToken)
	adminDisksRec := httptest.NewRecorder()
	router.ServeHTTP(adminDisksRec, adminDisksReq)
	require.Equal(t, http.StatusOK, adminDisksRec.Code)
	require.NoError(t, json.Unmarshal(adminDisksRec.Body.Bytes(), &disksResp))
	require.Equal(t, diskPath, disksResp.Disks[0].Path)
}

func TestDiskMasking_MonitoringOverviewForRegularUser(t *testing.T) {
	router, service, volumeService, diskPath := setupTestRouter(t)

	adminLogin, err := service.Login("admin@example.com", "adminpass1")
	require.NoError(t, err)

	patchReq := httptest.NewRequest(
		http.MethodPatch,
		"/api/admin/settings",
		bytes.NewReader([]byte(`{"mask_disk_names":true}`)),
	)
	patchReq.Header.Set("Content-Type", "application/json")
	patchReq.Header.Set("Authorization", "Bearer "+adminLogin.AccessToken)
	patchRec := httptest.NewRecorder()
	router.ServeHTTP(patchRec, patchReq)
	require.Equal(t, http.StatusOK, patchRec.Code)

	member, err := service.CreateUser("viewer@example.com", "password123", auth.RoleUser)
	require.NoError(t, err)

	vol, err := volumeService.Create(
		&auth.Claims{UserID: member.ID, Role: auth.RoleUser},
		volume.CreateVolumeInput{
			Name:       "Photos",
			DiskPath:   diskPath,
			QuotaBytes: 50 * 1024 * 1024 * 1024,
			Filters:    volume.Filters{Mode: volume.FilterModeAllow, Extensions: []string{".png"}},
		},
	)
	require.NoError(t, err)

	userLogin, err := service.Login("viewer@example.com", "password123")
	require.NoError(t, err)

	overviewReq := httptest.NewRequest(http.MethodGet, "/api/monitoring/overview", nil)
	overviewReq.Header.Set("Authorization", "Bearer "+userLogin.AccessToken)
	overviewRec := httptest.NewRecorder()
	router.ServeHTTP(overviewRec, overviewReq)
	require.Equal(t, http.StatusOK, overviewRec.Code)

	var overview monitoring.OverviewResponse
	require.NoError(t, json.Unmarshal(overviewRec.Body.Bytes(), &overview))
	require.NotEmpty(t, overview.Disks)
	require.Empty(t, overview.Disks[0].Path)
	require.Equal(t, "Storage 1", overview.Disks[0].Label)

	found := false
	for _, summary := range overview.Volumes {
		if summary.ID == vol.ID.String() {
			found = true
			require.Empty(t, summary.DiskPath)
		}
	}
	require.True(t, found)
}

func TestDiskMasking_PatchResponseIsMasked(t *testing.T) {
	router, service, volumeService, diskPath := setupTestRouter(t)

	adminLogin, err := service.Login("admin@example.com", "adminpass1")
	require.NoError(t, err)
	patchSettingsReq := httptest.NewRequest(
		http.MethodPatch,
		"/api/admin/settings",
		bytes.NewReader([]byte(`{"mask_disk_names":true}`)),
	)
	patchSettingsReq.Header.Set("Content-Type", "application/json")
	patchSettingsReq.Header.Set("Authorization", "Bearer "+adminLogin.AccessToken)
	patchSettingsRec := httptest.NewRecorder()
	router.ServeHTTP(patchSettingsRec, patchSettingsReq)
	require.Equal(t, http.StatusOK, patchSettingsRec.Code)

	member, err := service.CreateUser("editor@example.com", "password123", auth.RoleUser)
	require.NoError(t, err)

	vol, err := volumeService.Create(
		&auth.Claims{UserID: member.ID, Role: auth.RoleUser},
		volume.CreateVolumeInput{
			Name:       "Photos",
			DiskPath:   diskPath,
			QuotaBytes: 0,
			Filters:    volume.Filters{},
		},
	)
	require.NoError(t, err)

	userLogin, err := service.Login("editor@example.com", "password123")
	require.NoError(t, err)

	newName := "Renamed"
	patchBody, err := json.Marshal(map[string]string{"name": newName})
	require.NoError(t, err)
	patchReq := httptest.NewRequest(http.MethodPatch, "/api/volumes/"+vol.ID.String(), bytes.NewReader(patchBody))
	patchReq.Header.Set("Content-Type", "application/json")
	patchReq.Header.Set("Authorization", "Bearer "+userLogin.AccessToken)
	patchRec := httptest.NewRecorder()
	router.ServeHTTP(patchRec, patchReq)
	require.Equal(t, http.StatusOK, patchRec.Code)

	var patched struct {
		Name     string `json:"name"`
		DiskPath string `json:"disk_path"`
		RootPath string `json:"root_path"`
	}
	require.NoError(t, json.Unmarshal(patchRec.Body.Bytes(), &patched))
	require.Equal(t, newName, patched.Name)
	require.Empty(t, patched.DiskPath)
	require.Empty(t, patched.RootPath)
}
