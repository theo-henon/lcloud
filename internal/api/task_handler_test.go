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
