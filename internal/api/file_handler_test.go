package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileMoveAndRename(t *testing.T) {
	router, _, _, _, storageRoot := setupTestRouter(t)
	token := loginAsAdmin(t, router)
	volumeID := createVolume(t, router, token, storageRoot)
	uploadPNG(t, router, token, volumeID, "photo.png")

	moveBody := []byte(`{"from_path":"photo.png","to_path":"images/photo.png"}`)
	moveReq := httptest.NewRequest(http.MethodPatch, "/api/volumes/"+volumeID+"/files/move", bytes.NewReader(moveBody))
	moveReq.Header.Set("Content-Type", "application/json")
	moveReq.Header.Set("Authorization", "Bearer "+token)
	moveRec := httptest.NewRecorder()
	router.ServeHTTP(moveRec, moveReq)
	require.Equal(t, http.StatusOK, moveRec.Code)

	var moved map[string]any
	require.NoError(t, json.Unmarshal(moveRec.Body.Bytes(), &moved))
	require.Equal(t, "images/photo.png", moved["path"])

	renameBody := []byte(`{"path":"images/photo.png","new_name":"cover.png"}`)
	renameReq := httptest.NewRequest(http.MethodPatch, "/api/volumes/"+volumeID+"/files/rename", bytes.NewReader(renameBody))
	renameReq.Header.Set("Content-Type", "application/json")
	renameReq.Header.Set("Authorization", "Bearer "+token)
	renameRec := httptest.NewRecorder()
	router.ServeHTTP(renameRec, renameReq)
	require.Equal(t, http.StatusOK, renameRec.Code)

	var renamed map[string]any
	require.NoError(t, json.Unmarshal(renameRec.Body.Bytes(), &renamed))
	require.Equal(t, "images/cover.png", renamed["path"])
}

func TestFileContentInlineDisposition(t *testing.T) {
	router, _, _, _, storageRoot := setupTestRouter(t)
	token := loginAsAdmin(t, router)
	volumeID := createVolume(t, router, token, storageRoot)
	uploadPNG(t, router, token, volumeID, "inline.png")

	req := httptest.NewRequest(http.MethodGet, "/api/volumes/"+volumeID+"/files/content?path=inline.png&disposition=inline", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Header().Get("Content-Disposition"), "inline")
}

func TestFileMoveNotAFile(t *testing.T) {
	router, _, _, _, storageRoot := setupTestRouter(t)
	token := loginAsAdmin(t, router)
	volumeID := createVolume(t, router, token, storageRoot)

	mkdirBody := []byte(`{"path":"folder"}`)
	mkdirReq := httptest.NewRequest(http.MethodPost, "/api/volumes/"+volumeID+"/files/directories", bytes.NewReader(mkdirBody))
	mkdirReq.Header.Set("Content-Type", "application/json")
	mkdirReq.Header.Set("Authorization", "Bearer "+token)
	mkdirRec := httptest.NewRecorder()
	router.ServeHTTP(mkdirRec, mkdirReq)
	require.Equal(t, http.StatusCreated, mkdirRec.Code)

	moveBody := []byte(`{"from_path":"folder","to_path":"folder2"}`)
	moveReq := httptest.NewRequest(http.MethodPatch, "/api/volumes/"+volumeID+"/files/move", bytes.NewReader(moveBody))
	moveReq.Header.Set("Content-Type", "application/json")
	moveReq.Header.Set("Authorization", "Bearer "+token)
	moveRec := httptest.NewRecorder()
	router.ServeHTTP(moveRec, moveReq)
	require.Equal(t, http.StatusUnprocessableEntity, moveRec.Code)
}
