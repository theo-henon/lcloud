package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/task"
	"github.com/theo-henon/lcloud/internal/volume"
)

func TestTaskCRUD(t *testing.T) {
	router, service, _, storageRoot := setupTestRouter(t)

	login, err := service.Login("admin@example.com", "adminpass1")
	require.NoError(t, err)

	volBody, err := json.Marshal(map[string]any{
		"name":         "Photos",
		"disk_path":    storageRoot,
		"quota_bytes":  50 * 1024 * 1024 * 1024,
		"filters":      map[string]any{},
	})
	require.NoError(t, err)
	volReq := httptest.NewRequest(http.MethodPost, "/api/volumes", bytes.NewReader(volBody))
	volReq.Header.Set("Content-Type", "application/json")
	volReq.Header.Set("Authorization", "Bearer "+login.AccessToken)
	volRec := httptest.NewRecorder()
	router.ServeHTTP(volRec, volReq)
	require.Equal(t, http.StatusCreated, volRec.Code)

	var vol volume.Volume
	require.NoError(t, json.Unmarshal(volRec.Body.Bytes(), &vol))

	createBody, err := json.Marshal(map[string]any{
		"name":          "Delete old photos",
		"macro":         task.MacroDeleteOldFiles,
		"scope":         task.ScopeVolume,
		"volume_id":     vol.ID.String(),
		"parameters":    map[string]any{"days": 90},
		"schedule_type": task.ScheduleTypeCron,
		"schedule":      "0 2 * * 0",
		"enabled":       true,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+login.AccessToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var created task.TaskView
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	require.Equal(t, task.MacroDeleteOldFiles, created.Macro)

	listReq := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	listReq.Header.Set("Authorization", "Bearer "+login.AccessToken)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)
	require.Equal(t, http.StatusOK, listRec.Code)

	runReq := httptest.NewRequest(http.MethodPost, "/api/tasks/"+created.ID.String()+"/run?dry_run=true", nil)
	runReq.Header.Set("Authorization", "Bearer "+login.AccessToken)
	runRec := httptest.NewRecorder()
	router.ServeHTTP(runRec, runReq)
	require.Equal(t, http.StatusAccepted, runRec.Code)

	time.Sleep(200 * time.Millisecond)

	runsReq := httptest.NewRequest(http.MethodGet, "/api/tasks/"+created.ID.String()+"/runs", nil)
	runsReq.Header.Set("Authorization", "Bearer "+login.AccessToken)
	runsRec := httptest.NewRecorder()
	router.ServeHTTP(runsRec, runsReq)
	require.Equal(t, http.StatusOK, runsRec.Code)
}

func TestTaskGlobalForbiddenForUser(t *testing.T) {
	router, service, _, _ := setupTestRouter(t)

	_, err := service.CreateUser("user@example.com", "password123", auth.RoleUser)
	require.NoError(t, err)
	login, err := service.Login("user@example.com", "password123")
	require.NoError(t, err)

	body, err := json.Marshal(map[string]any{
		"name":          "Global stats",
		"macro":         task.MacroComputeStats,
		"scope":         task.ScopeGlobal,
		"parameters":    map[string]any{},
		"schedule_type": task.ScheduleTypeInterval,
		"schedule":      "24h",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+login.AccessToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestTaskRunDisabled(t *testing.T) {
	router, service, _, storageRoot := setupTestRouter(t)

	login, err := service.Login("admin@example.com", "adminpass1")
	require.NoError(t, err)

	volBody, err := json.Marshal(map[string]any{
		"name":        "Photos",
		"disk_path":   storageRoot,
		"quota_bytes": 50 * 1024 * 1024 * 1024,
		"filters":     map[string]any{},
	})
	require.NoError(t, err)
	volReq := httptest.NewRequest(http.MethodPost, "/api/volumes", bytes.NewReader(volBody))
	volReq.Header.Set("Content-Type", "application/json")
	volReq.Header.Set("Authorization", "Bearer "+login.AccessToken)
	volRec := httptest.NewRecorder()
	router.ServeHTTP(volRec, volReq)
	require.Equal(t, http.StatusCreated, volRec.Code)

	var vol volume.Volume
	require.NoError(t, json.Unmarshal(volRec.Body.Bytes(), &vol))

	createBody, err := json.Marshal(map[string]any{
		"name":          "Disabled cleanup",
		"macro":         task.MacroDeleteOldFiles,
		"scope":         task.ScopeVolume,
		"volume_id":     vol.ID.String(),
		"parameters":    map[string]any{"days": 90},
		"schedule_type": task.ScheduleTypeCron,
		"schedule":      "0 2 * * 0",
		"enabled":       false,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+login.AccessToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var created task.TaskView
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))

	runReq := httptest.NewRequest(http.MethodPost, "/api/tasks/"+created.ID.String()+"/run", nil)
	runReq.Header.Set("Authorization", "Bearer "+login.AccessToken)
	runRec := httptest.NewRecorder()
	router.ServeHTTP(runRec, runReq)
	require.Equal(t, http.StatusUnprocessableEntity, runRec.Code)
}

func TestTaskListIsolatedBetweenUsers(t *testing.T) {
	router, service, volumeService, diskPath := setupTestRouter(t)

	userA, err := service.CreateUser("alice@example.com", "password123", auth.RoleUser)
	require.NoError(t, err)
	_, err = service.CreateUser("bob@example.com", "password123", auth.RoleUser)
	require.NoError(t, err)

	vol, err := volumeService.Create(
		&auth.Claims{UserID: userA.ID, Role: auth.RoleUser},
		volume.CreateVolumeInput{Name: "Alice Photos", DiskPath: diskPath},
	)
	require.NoError(t, err)

	aliceLogin, err := service.Login("alice@example.com", "password123")
	require.NoError(t, err)

	createBody, err := json.Marshal(map[string]any{
		"name":          "Alice cleanup",
		"macro":         task.MacroDeleteOldFiles,
		"scope":         task.ScopeVolume,
		"volume_id":     vol.ID.String(),
		"parameters":    map[string]any{"days": 90},
		"schedule_type": task.ScheduleTypeCron,
		"schedule":      "0 2 * * 0",
		"enabled":       true,
	})
	require.NoError(t, err)

	createReq := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+aliceLogin.AccessToken)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	bobLogin, err := service.Login("bob@example.com", "password123")
	require.NoError(t, err)

	listReq := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	listReq.Header.Set("Authorization", "Bearer "+bobLogin.AccessToken)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)
	require.Equal(t, http.StatusOK, listRec.Code)

	var resp struct {
		Tasks []task.TaskView `json:"tasks"`
	}
	require.NoError(t, json.Unmarshal(listRec.Body.Bytes(), &resp))
	require.Empty(t, resp.Tasks)
}

func TestTaskCreateForbiddenOnForeignVolume(t *testing.T) {
	router, service, volumeService, diskPath := setupTestRouter(t)

	userA, err := service.CreateUser("alice@example.com", "password123", auth.RoleUser)
	require.NoError(t, err)
	_, err = service.CreateUser("bob@example.com", "password123", auth.RoleUser)
	require.NoError(t, err)

	vol, err := volumeService.Create(
		&auth.Claims{UserID: userA.ID, Role: auth.RoleUser},
		volume.CreateVolumeInput{Name: "Alice Photos", DiskPath: diskPath},
	)
	require.NoError(t, err)

	bobLogin, err := service.Login("bob@example.com", "password123")
	require.NoError(t, err)

	createBody, err := json.Marshal(map[string]any{
		"name":          "Bob on Alice volume",
		"macro":         task.MacroDeleteOldFiles,
		"scope":         task.ScopeVolume,
		"volume_id":     vol.ID.String(),
		"parameters":    map[string]any{"days": 90},
		"schedule_type": task.ScheduleTypeCron,
		"schedule":      "0 2 * * 0",
		"enabled":       true,
	})
	require.NoError(t, err)

	createReq := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+bobLogin.AccessToken)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusForbidden, createRec.Code)
}

func TestTaskListIncludesOwnerEmailForAdmin(t *testing.T) {
	router, service, volumeService, diskPath := setupTestRouter(t)

	user, err := service.CreateUser("creator@example.com", "password123", auth.RoleUser)
	require.NoError(t, err)

	vol, err := volumeService.Create(
		&auth.Claims{UserID: user.ID, Role: auth.RoleUser},
		volume.CreateVolumeInput{Name: "Photos", DiskPath: diskPath},
	)
	require.NoError(t, err)

	userLogin, err := service.Login("creator@example.com", "password123")
	require.NoError(t, err)

	createBody, err := json.Marshal(map[string]any{
		"name":          "Weekly cleanup",
		"macro":         task.MacroDeleteOldFiles,
		"scope":         task.ScopeVolume,
		"volume_id":     vol.ID.String(),
		"parameters":    map[string]any{"days": 90},
		"schedule_type": task.ScheduleTypeCron,
		"schedule":      "0 2 * * 0",
		"enabled":       true,
	})
	require.NoError(t, err)

	createReq := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+userLogin.AccessToken)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	adminLogin, err := service.Login("admin@example.com", "adminpass1")
	require.NoError(t, err)

	listReq := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	listReq.Header.Set("Authorization", "Bearer "+adminLogin.AccessToken)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)
	require.Equal(t, http.StatusOK, listRec.Code)

	var resp struct {
		Tasks []struct {
			Name       string `json:"name"`
			OwnerEmail string `json:"owner_email"`
		} `json:"tasks"`
	}
	require.NoError(t, json.Unmarshal(listRec.Body.Bytes(), &resp))
	require.NotEmpty(t, resp.Tasks)
	require.Equal(t, "creator@example.com", resp.Tasks[0].OwnerEmail)
}

func TestTaskListOmitsOwnerEmailForRegularUser(t *testing.T) {
	router, service, volumeService, diskPath := setupTestRouter(t)

	user, err := service.CreateUser("creator@example.com", "password123", auth.RoleUser)
	require.NoError(t, err)

	vol, err := volumeService.Create(
		&auth.Claims{UserID: user.ID, Role: auth.RoleUser},
		volume.CreateVolumeInput{Name: "Photos", DiskPath: diskPath},
	)
	require.NoError(t, err)

	userLogin, err := service.Login("creator@example.com", "password123")
	require.NoError(t, err)

	createBody, err := json.Marshal(map[string]any{
		"name":          "Weekly cleanup",
		"macro":         task.MacroDeleteOldFiles,
		"scope":         task.ScopeVolume,
		"volume_id":     vol.ID.String(),
		"parameters":    map[string]any{"days": 90},
		"schedule_type": task.ScheduleTypeCron,
		"schedule":      "0 2 * * 0",
		"enabled":       true,
	})
	require.NoError(t, err)

	createReq := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+userLogin.AccessToken)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	listReq := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	listReq.Header.Set("Authorization", "Bearer "+userLogin.AccessToken)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)
	require.Equal(t, http.StatusOK, listRec.Code)

	var raw map[string]any
	require.NoError(t, json.Unmarshal(listRec.Body.Bytes(), &raw))
	tasks := raw["tasks"].([]any)
	item := tasks[0].(map[string]any)
	_, hasOwnerEmail := item["owner_email"]
	require.False(t, hasOwnerEmail)
}

func TestVolumeListFilterByOwnerForAdmin(t *testing.T) {
	router, service, volumeService, diskPath := setupTestRouter(t)

	user, err := service.CreateUser("creator@example.com", "password123", auth.RoleUser)
	require.NoError(t, err)

	_, err = volumeService.Create(
		&auth.Claims{UserID: user.ID, Role: auth.RoleUser},
		volume.CreateVolumeInput{Name: "User Vol", DiskPath: diskPath},
	)
	require.NoError(t, err)

	adminLogin, err := service.Login("admin@example.com", "adminpass1")
	require.NoError(t, err)

	adminVolBody, err := json.Marshal(map[string]any{
		"name":        "Admin Vol",
		"disk_path":   diskPath,
		"quota_bytes": 0,
		"filters":     map[string]any{},
	})
	require.NoError(t, err)
	adminVolReq := httptest.NewRequest(http.MethodPost, "/api/volumes", bytes.NewReader(adminVolBody))
	adminVolReq.Header.Set("Content-Type", "application/json")
	adminVolReq.Header.Set("Authorization", "Bearer "+adminLogin.AccessToken)
	adminVolRec := httptest.NewRecorder()
	router.ServeHTTP(adminVolRec, adminVolReq)
	require.Equal(t, http.StatusCreated, adminVolRec.Code)

	listAllReq := httptest.NewRequest(http.MethodGet, "/api/volumes", nil)
	listAllReq.Header.Set("Authorization", "Bearer "+adminLogin.AccessToken)
	listAllRec := httptest.NewRecorder()
	router.ServeHTTP(listAllRec, listAllReq)
	require.Equal(t, http.StatusOK, listAllRec.Code)

	var allResp struct {
		Volumes []struct {
			Name string `json:"name"`
		} `json:"volumes"`
	}
	require.NoError(t, json.Unmarshal(listAllRec.Body.Bytes(), &allResp))
	require.GreaterOrEqual(t, len(allResp.Volumes), 2)

	filterReq := httptest.NewRequest(
		http.MethodGet,
		"/api/volumes?owner_id="+user.ID.String(),
		nil,
	)
	filterReq.Header.Set("Authorization", "Bearer "+adminLogin.AccessToken)
	filterRec := httptest.NewRecorder()
	router.ServeHTTP(filterRec, filterReq)
	require.Equal(t, http.StatusOK, filterRec.Code)

	var filteredResp struct {
		Volumes []struct {
			Name string `json:"name"`
		} `json:"volumes"`
	}
	require.NoError(t, json.Unmarshal(filterRec.Body.Bytes(), &filteredResp))
	require.Len(t, filteredResp.Volumes, 1)
	require.Equal(t, "User Vol", filteredResp.Volumes[0].Name)
}

func TestTaskListFilterByOwnerForAdmin(t *testing.T) {
	router, service, volumeService, diskPath := setupTestRouter(t)

	user, err := service.CreateUser("creator@example.com", "password123", auth.RoleUser)
	require.NoError(t, err)

	vol, err := volumeService.Create(
		&auth.Claims{UserID: user.ID, Role: auth.RoleUser},
		volume.CreateVolumeInput{Name: "Photos", DiskPath: diskPath},
	)
	require.NoError(t, err)

	userLogin, err := service.Login("creator@example.com", "password123")
	require.NoError(t, err)

	createBody, err := json.Marshal(map[string]any{
		"name":          "User task",
		"macro":         task.MacroDeleteOldFiles,
		"scope":         task.ScopeVolume,
		"volume_id":     vol.ID.String(),
		"parameters":    map[string]any{"days": 90},
		"schedule_type": task.ScheduleTypeCron,
		"schedule":      "0 2 * * 0",
		"enabled":       true,
	})
	require.NoError(t, err)

	createReq := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+userLogin.AccessToken)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	adminLogin, err := service.Login("admin@example.com", "adminpass1")
	require.NoError(t, err)

	filterReq := httptest.NewRequest(
		http.MethodGet,
		"/api/tasks?owner_id="+user.ID.String(),
		nil,
	)
	filterReq.Header.Set("Authorization", "Bearer "+adminLogin.AccessToken)
	filterRec := httptest.NewRecorder()
	router.ServeHTTP(filterRec, filterReq)
	require.Equal(t, http.StatusOK, filterRec.Code)

	var resp struct {
		Tasks []struct {
			Name string `json:"name"`
		} `json:"tasks"`
	}
	require.NoError(t, json.Unmarshal(filterRec.Body.Bytes(), &resp))
	require.Len(t, resp.Tasks, 1)
	require.Equal(t, "User task", resp.Tasks[0].Name)
}
