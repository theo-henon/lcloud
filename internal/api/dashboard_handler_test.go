package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/dashboard"
)

func TestDashboardUnauthorized(t *testing.T) {
	router, _ := setupTestRouterLegacy(t)

	req := httptest.NewRequest(http.MethodGet, "/api/users/me/dashboard", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestDashboardGetDefault(t *testing.T) {
	router, service := setupTestRouterLegacy(t)

	login, err := service.Login("admin@example.com", "adminpass1")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/users/me/dashboard", nil)
	req.Header.Set("Authorization", "Bearer "+login.AccessToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var resp dashboard.DashboardResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Nil(t, resp.UpdatedAt)
	require.Equal(t, dashboard.DefaultLayout(auth.RoleAdmin), resp.Layout)
	require.Len(t, resp.Catalog, 7)
}

func TestDashboardPatchRoundTrip(t *testing.T) {
	router, service := setupTestRouterLegacy(t)

	login, err := service.Login("admin@example.com", "adminpass1")
	require.NoError(t, err)

	layout := dashboard.Layout{
		Version: dashboard.LayoutVersion,
		Widgets: []dashboard.WidgetPlacement{
			{ID: "w1", Type: "welcome", X: 0, Y: 0, W: 4, H: 1},
			{ID: "w2", Type: "quick-actions", X: 4, Y: 0, W: 4, H: 1},
		},
	}
	body, err := json.Marshal(map[string]any{"layout": layout})
	require.NoError(t, err)

	patchReq := httptest.NewRequest(http.MethodPatch, "/api/users/me/dashboard", bytes.NewReader(body))
	patchReq.Header.Set("Content-Type", "application/json")
	patchReq.Header.Set("Authorization", "Bearer "+login.AccessToken)
	patchRec := httptest.NewRecorder()
	router.ServeHTTP(patchRec, patchReq)
	require.Equal(t, http.StatusOK, patchRec.Code)

	getReq := httptest.NewRequest(http.MethodGet, "/api/users/me/dashboard", nil)
	getReq.Header.Set("Authorization", "Bearer "+login.AccessToken)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)
	require.Equal(t, http.StatusOK, getRec.Code)

	var resp dashboard.DashboardResponse
	require.NoError(t, json.Unmarshal(getRec.Body.Bytes(), &resp))
	require.NotNil(t, resp.UpdatedAt)
	require.Equal(t, layout, resp.Layout)
}

func TestDashboardUserCannotSaveAdminWidget(t *testing.T) {
	router, service := setupTestRouterLegacy(t)

	_, err := service.CreateUser("user@example.com", "password123", auth.RoleUser)
	require.NoError(t, err)

	login, err := service.Login("user@example.com", "password123")
	require.NoError(t, err)

	layout := dashboard.Layout{
		Version: dashboard.LayoutVersion,
		Widgets: []dashboard.WidgetPlacement{
			{ID: "w1", Type: "pending-deletions", X: 0, Y: 0, W: 6, H: 2},
		},
	}
	body, err := json.Marshal(map[string]any{"layout": layout})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPatch, "/api/users/me/dashboard", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+login.AccessToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)

	var errResp struct {
		Code string `json:"code"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &errResp))
	require.Equal(t, "WIDGET_NOT_ALLOWED", errResp.Code)
}

func TestDashboardReset(t *testing.T) {
	router, service := setupTestRouterLegacy(t)

	login, err := service.Login("admin@example.com", "adminpass1")
	require.NoError(t, err)

	layout := dashboard.Layout{
		Version: dashboard.LayoutVersion,
		Widgets: []dashboard.WidgetPlacement{
			{ID: "w1", Type: "welcome", X: 0, Y: 0, W: 4, H: 1},
		},
	}
	body, err := json.Marshal(map[string]any{"layout": layout})
	require.NoError(t, err)

	patchReq := httptest.NewRequest(http.MethodPatch, "/api/users/me/dashboard", bytes.NewReader(body))
	patchReq.Header.Set("Content-Type", "application/json")
	patchReq.Header.Set("Authorization", "Bearer "+login.AccessToken)
	patchRec := httptest.NewRecorder()
	router.ServeHTTP(patchRec, patchReq)
	require.Equal(t, http.StatusOK, patchRec.Code)

	delReq := httptest.NewRequest(http.MethodDelete, "/api/users/me/dashboard", nil)
	delReq.Header.Set("Authorization", "Bearer "+login.AccessToken)
	delRec := httptest.NewRecorder()
	router.ServeHTTP(delRec, delReq)
	require.Equal(t, http.StatusNoContent, delRec.Code)

	getReq := httptest.NewRequest(http.MethodGet, "/api/users/me/dashboard", nil)
	getReq.Header.Set("Authorization", "Bearer "+login.AccessToken)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)
	require.Equal(t, http.StatusOK, getRec.Code)

	var resp dashboard.DashboardResponse
	require.NoError(t, json.Unmarshal(getRec.Body.Bytes(), &resp))
	require.Nil(t, resp.UpdatedAt)
}
