package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNotificationUnauthorized(t *testing.T) {
	router, _, _, _, _ := setupTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestNotificationListEmpty(t *testing.T) {
	router, authSvc, _, _, _ := setupTestRouter(t)

	login, err := authSvc.Login("admin@example.com", "adminpass1")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+login.AccessToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var listResp struct {
		Notifications []any `json:"notifications"`
		Total         int64 `json:"total"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &listResp))
	require.Equal(t, int64(0), listResp.Total)
	require.Empty(t, listResp.Notifications)
}

func TestNotificationMarkReadDismissCrossUser(t *testing.T) {
	router, authSvc, _, notifSvc, _ := setupTestRouter(t)

	adminLogin, err := authSvc.Login("admin@example.com", "adminpass1")
	require.NoError(t, err)

	_, err = authSvc.CreateUser("notify@example.com", "userpass12", "user")
	require.NoError(t, err)

	item, err := notifSvc.Seed(context.Background(), adminLogin.User.ID, "Test alert", "api-seed-1")
	require.NoError(t, err)

	userLogin, err := authSvc.Login("notify@example.com", "userpass12")
	require.NoError(t, err)

	patchReq := httptest.NewRequest(http.MethodPatch, "/api/notifications/"+item.ID.String()+"/read", nil)
	patchReq.Header.Set("Authorization", "Bearer "+userLogin.AccessToken)
	patchRec := httptest.NewRecorder()
	router.ServeHTTP(patchRec, patchReq)
	require.Equal(t, http.StatusNotFound, patchRec.Code)

	delReq := httptest.NewRequest(http.MethodDelete, "/api/notifications/"+item.ID.String(), nil)
	delReq.Header.Set("Authorization", "Bearer "+userLogin.AccessToken)
	delRec := httptest.NewRecorder()
	router.ServeHTTP(delRec, delReq)
	require.Equal(t, http.StatusNotFound, delRec.Code)

	markReq := httptest.NewRequest(http.MethodPatch, "/api/notifications/"+item.ID.String()+"/read", nil)
	markReq.Header.Set("Authorization", "Bearer "+adminLogin.AccessToken)
	markRec := httptest.NewRecorder()
	router.ServeHTTP(markRec, markReq)
	require.Equal(t, http.StatusOK, markRec.Code)

	countReq := httptest.NewRequest(http.MethodGet, "/api/notifications/unread-count", nil)
	countReq.Header.Set("Authorization", "Bearer "+adminLogin.AccessToken)
	countRec := httptest.NewRecorder()
	router.ServeHTTP(countRec, countReq)
	require.Equal(t, http.StatusOK, countRec.Code)

	var countResp struct {
		Count int64 `json:"count"`
	}
	require.NoError(t, json.Unmarshal(countRec.Body.Bytes(), &countResp))
	require.Equal(t, int64(0), countResp.Count)

	readAllReq := httptest.NewRequest(http.MethodPost, "/api/notifications/read-all", nil)
	readAllReq.Header.Set("Authorization", "Bearer "+adminLogin.AccessToken)
	readAllRec := httptest.NewRecorder()
	router.ServeHTTP(readAllRec, readAllReq)
	require.Equal(t, http.StatusOK, readAllRec.Code)

	dismissReq := httptest.NewRequest(http.MethodDelete, "/api/notifications/"+item.ID.String(), nil)
	dismissReq.Header.Set("Authorization", "Bearer "+adminLogin.AccessToken)
	dismissRec := httptest.NewRecorder()
	router.ServeHTTP(dismissRec, dismissReq)
	require.Equal(t, http.StatusNoContent, dismissRec.Code)

	listReq := httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
	listReq.Header.Set("Authorization", "Bearer "+adminLogin.AccessToken)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)
	require.Equal(t, http.StatusOK, listRec.Code)

	var listResp struct {
		Total int64 `json:"total"`
	}
	require.NoError(t, json.Unmarshal(listRec.Body.Bytes(), &listResp))
	require.Equal(t, int64(0), listResp.Total)
}
