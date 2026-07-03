package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/theo-henon/lcloud/internal/volume"
)

func TestTrashListRestoreAndPurge(t *testing.T) {
	router, _, _, storageRoot := setupTestRouter(t)
	token := loginAsAdmin(t, router)
	volumeID := createVolume(t, router, token, storageRoot)
	uploadPNG(t, router, token, volumeID, "trash-me.png")

	delReq := httptest.NewRequest(http.MethodDelete, "/api/volumes/"+volumeID+"/files?path=trash-me.png", nil)
	delReq.Header.Set("Authorization", "Bearer "+token)
	delRec := httptest.NewRecorder()
	router.ServeHTTP(delRec, delReq)
	require.Equal(t, http.StatusNoContent, delRec.Code)

	listReq := httptest.NewRequest(http.MethodGet, "/api/volumes/"+volumeID+"/trash", nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)
	require.Equal(t, http.StatusOK, listRec.Code)

	var listing struct {
		Items []volume.TrashEntry `json:"items"`
		Total int                 `json:"total"`
	}
	require.NoError(t, json.Unmarshal(listRec.Body.Bytes(), &listing))
	require.Equal(t, 1, listing.Total)
	require.Equal(t, "trash-me.png", listing.Items[0].Name)
	fileID := listing.Items[0].ID

	restoreReq := httptest.NewRequest(http.MethodPost, "/api/volumes/"+volumeID+"/trash/"+fileID+"/restore", nil)
	restoreReq.Header.Set("Authorization", "Bearer "+token)
	restoreRec := httptest.NewRecorder()
	router.ServeHTTP(restoreRec, restoreReq)
	require.Equal(t, http.StatusOK, restoreRec.Code)

	listReq2 := httptest.NewRequest(http.MethodGet, "/api/volumes/"+volumeID+"/files?path=.", nil)
	listReq2.Header.Set("Authorization", "Bearer "+token)
	listRec2 := httptest.NewRecorder()
	router.ServeHTTP(listRec2, listReq2)
	require.Equal(t, http.StatusOK, listRec2.Code)

	// Re-trash and permanently delete
	delReq2 := httptest.NewRequest(http.MethodDelete, "/api/volumes/"+volumeID+"/files?path=trash-me.png", nil)
	delReq2.Header.Set("Authorization", "Bearer "+token)
	delRec2 := httptest.NewRecorder()
	router.ServeHTTP(delRec2, delReq2)
	require.Equal(t, http.StatusNoContent, delRec2.Code)

	listReq3 := httptest.NewRequest(http.MethodGet, "/api/volumes/"+volumeID+"/trash", nil)
	listReq3.Header.Set("Authorization", "Bearer "+token)
	listRec3 := httptest.NewRecorder()
	router.ServeHTTP(listRec3, listReq3)
	require.NoError(t, json.Unmarshal(listRec3.Body.Bytes(), &listing))
	require.Equal(t, 1, listing.Total)
	fileID = listing.Items[0].ID

	purgeReq := httptest.NewRequest(http.MethodDelete, "/api/volumes/"+volumeID+"/trash/"+fileID, nil)
	purgeReq.Header.Set("Authorization", "Bearer "+token)
	purgeRec := httptest.NewRecorder()
	router.ServeHTTP(purgeRec, purgeReq)
	require.Equal(t, http.StatusNoContent, purgeRec.Code)

	listReq4 := httptest.NewRequest(http.MethodGet, "/api/volumes/"+volumeID+"/trash", nil)
	listReq4.Header.Set("Authorization", "Bearer "+token)
	listRec4 := httptest.NewRecorder()
	router.ServeHTTP(listRec4, listReq4)
	require.NoError(t, json.Unmarshal(listRec4.Body.Bytes(), &listing))
	require.Equal(t, 0, listing.Total)
}

func TestDeleteFileSoftDeletesToTrash(t *testing.T) {
	router, _, _, storageRoot := setupTestRouter(t)
	token := loginAsAdmin(t, router)
	volumeID := createVolume(t, router, token, storageRoot)
	uploadPNG(t, router, token, volumeID, "soft.png")

	delReq := httptest.NewRequest(http.MethodDelete, "/api/volumes/"+volumeID+"/files?path=soft.png", nil)
	delReq.Header.Set("Authorization", "Bearer "+token)
	delRec := httptest.NewRecorder()
	router.ServeHTTP(delRec, delReq)
	require.Equal(t, http.StatusNoContent, delRec.Code)

	volRoot := filepath.Join(storageRoot, volumeID)
	userdataFile := filepath.Join(volRoot, volume.UserdataDir, "soft.png")
	_, err := os.Stat(userdataFile)
	require.True(t, os.IsNotExist(err))

	trashDirs, err := os.ReadDir(filepath.Join(volRoot, volume.TrashDir))
	require.NoError(t, err)
	require.Len(t, trashDirs, 1)
}

func TestEmptyTrash(t *testing.T) {
	router, _, _, storageRoot := setupTestRouter(t)
	token := loginAsAdmin(t, router)
	volumeID := createVolume(t, router, token, storageRoot)
	uploadPNG(t, router, token, volumeID, "a.png")
	uploadPNG(t, router, token, volumeID, "b.png")

	for _, name := range []string{"a.png", "b.png"} {
		req := httptest.NewRequest(http.MethodDelete, "/api/volumes/"+volumeID+"/files?path="+name, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusNoContent, rec.Code)
	}

	emptyReq := httptest.NewRequest(http.MethodDelete, "/api/volumes/"+volumeID+"/trash", nil)
	emptyReq.Header.Set("Authorization", "Bearer "+token)
	emptyRec := httptest.NewRecorder()
	router.ServeHTTP(emptyRec, emptyReq)
	require.Equal(t, http.StatusOK, emptyRec.Code)

	var result map[string]any
	require.NoError(t, json.Unmarshal(emptyRec.Body.Bytes(), &result))
	require.EqualValues(t, 2, result["purged"])

	listReq := httptest.NewRequest(http.MethodGet, "/api/volumes/"+volumeID+"/trash", nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)

	var listing struct {
		Total int `json:"total"`
	}
	require.NoError(t, json.Unmarshal(listRec.Body.Bytes(), &listing))
	require.Equal(t, 0, listing.Total)
}

func TestTrashRestoreRejectsPathTraversalFileID(t *testing.T) {
	router, _, _, storageRoot := setupTestRouter(t)
	token := loginAsAdmin(t, router)
	volumeID := createVolume(t, router, token, storageRoot)
	_ = storageRoot

	restoreReq := httptest.NewRequest(
		http.MethodPost,
		"/api/volumes/"+volumeID+"/trash/x..y/restore",
		nil,
	)
	restoreReq.Header.Set("Authorization", "Bearer "+token)
	restoreRec := httptest.NewRecorder()
	router.ServeHTTP(restoreRec, restoreReq)
	require.Equal(t, http.StatusUnprocessableEntity, restoreRec.Code)

	purgeReq := httptest.NewRequest(
		http.MethodDelete,
		"/api/volumes/"+volumeID+"/trash/x..y",
		nil,
	)
	purgeReq.Header.Set("Authorization", "Bearer "+token)
	purgeRec := httptest.NewRecorder()
	router.ServeHTTP(purgeRec, purgeReq)
	require.Equal(t, http.StatusUnprocessableEntity, purgeRec.Code)
}
