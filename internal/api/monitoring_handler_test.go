package api

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/monitoring"
)

func loginAsAdmin(t *testing.T, router http.Handler) string {
	t.Helper()
	loginBody := []byte(`{"email":"admin@example.com","password":"adminpass1"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var loginResp auth.LoginResult
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &loginResp))
	return loginResp.AccessToken
}

func createVolume(t *testing.T, router http.Handler, token, diskPath string) string {
	t.Helper()
	body := []byte(`{"name":"Photos","disk_path":"` + diskPath + `","quota_bytes":53687091200,"filters":{"mode":"allow","extensions":[".png"]}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/volumes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	return resp["id"].(string)
}

func uploadPNG(t *testing.T, router http.Handler, token, volumeID, filename string) {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", filename)
	require.NoError(t, err)
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	require.NoError(t, png.Encode(part, img))
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/api/volumes/"+volumeID+"/files?path=.", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)
}

func TestMonitoringOverview(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := loginAsAdmin(t, router)

	req := httptest.NewRequest(http.MethodGet, "/api/monitoring/overview", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var overview monitoring.OverviewResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &overview))
	require.NotEmpty(t, overview.Disks)
}

func TestMonitoringVolumeStatsAndSearch(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := loginAsAdmin(t, router)

	disksReq := httptest.NewRequest(http.MethodGet, "/api/disks", nil)
	disksReq.Header.Set("Authorization", "Bearer "+token)
	disksRec := httptest.NewRecorder()
	router.ServeHTTP(disksRec, disksReq)
	require.Equal(t, http.StatusOK, disksRec.Code)

	var disksResp struct {
		Disks []struct {
			Path string `json:"path"`
		} `json:"disks"`
	}
	require.NoError(t, json.Unmarshal(disksRec.Body.Bytes(), &disksResp))
	require.NotEmpty(t, disksResp.Disks)

	volumeID := createVolume(t, router, token, disksResp.Disks[0].Path)
	uploadPNG(t, router, token, volumeID, "vacation.png")

	statsReq := httptest.NewRequest(http.MethodGet, "/api/monitoring/volumes/"+volumeID+"/stats", nil)
	statsReq.Header.Set("Authorization", "Bearer "+token)
	statsRec := httptest.NewRecorder()
	router.ServeHTTP(statsRec, statsReq)
	require.Equal(t, http.StatusOK, statsRec.Code)

	var stats monitoring.VolumeStats
	require.NoError(t, json.Unmarshal(statsRec.Body.Bytes(), &stats))
	require.Equal(t, 1, stats.FileCount)
	require.Equal(t, monitoring.CategoryImages, stats.ByCategory[0].Category)

	searchReq := httptest.NewRequest(http.MethodGet, "/api/volumes/"+volumeID+"/search?q=vacation", nil)
	searchReq.Header.Set("Authorization", "Bearer "+token)
	searchRec := httptest.NewRecorder()
	router.ServeHTTP(searchRec, searchReq)
	require.Equal(t, http.StatusOK, searchRec.Code)

	var searchResp struct {
		Total   int `json:"total"`
		Results []struct {
			Name string `json:"name"`
		} `json:"results"`
	}
	require.NoError(t, json.Unmarshal(searchRec.Body.Bytes(), &searchResp))
	require.Equal(t, 1, searchResp.Total)
	require.Equal(t, "vacation.png", searchResp.Results[0].Name)

	globalSearchReq := httptest.NewRequest(http.MethodGet, "/api/search?q=vacation", nil)
	globalSearchReq.Header.Set("Authorization", "Bearer "+token)
	globalSearchRec := httptest.NewRecorder()
	router.ServeHTTP(globalSearchRec, globalSearchReq)
	require.Equal(t, http.StatusOK, globalSearchRec.Code)

	var globalSearchResp struct {
		Total   int `json:"total"`
		Results []struct {
			Name       string `json:"name"`
			VolumeName string `json:"volume_name"`
		} `json:"results"`
	}
	require.NoError(t, json.Unmarshal(globalSearchRec.Body.Bytes(), &globalSearchResp))
	require.Equal(t, 1, globalSearchResp.Total)
	require.Equal(t, "vacation.png", globalSearchResp.Results[0].Name)
	require.Equal(t, "Photos", globalSearchResp.Results[0].VolumeName)
}

func TestMonitoringForbiddenForOtherUserVolume(t *testing.T) {
	router, service := setupTestRouter(t)
	adminToken := loginAsAdmin(t, router)

	disksReq := httptest.NewRequest(http.MethodGet, "/api/disks", nil)
	disksReq.Header.Set("Authorization", "Bearer "+adminToken)
	disksRec := httptest.NewRecorder()
	router.ServeHTTP(disksRec, disksReq)
	var disksResp struct {
		Disks []struct {
			Path string `json:"path"`
		} `json:"disks"`
	}
	require.NoError(t, json.Unmarshal(disksRec.Body.Bytes(), &disksResp))

	volumeID := createVolume(t, router, adminToken, disksResp.Disks[0].Path)

	otherUser, err := service.CreateUser("other@example.com", "password123", auth.RoleUser)
	require.NoError(t, err)
	otherLogin, err := service.Login(otherUser.Email, "password123")
	require.NoError(t, err)

	statsReq := httptest.NewRequest(http.MethodGet, "/api/monitoring/volumes/"+volumeID+"/stats", nil)
	statsReq.Header.Set("Authorization", "Bearer "+otherLogin.AccessToken)
	statsRec := httptest.NewRecorder()
	router.ServeHTTP(statsRec, statsReq)
	require.Equal(t, http.StatusForbidden, statsRec.Code)

	_, err = uuid.Parse(volumeID)
	require.NoError(t, err)
}
