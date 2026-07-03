package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/volume"
)

func TestProtocolHandlers(t *testing.T) {
	router, authService, volumeService, storageRoot := setupTestRouter(t)

	login, err := authService.Login("admin@example.com", "adminpass1")
	require.NoError(t, err)

	adminClaims, err := authService.ValidateAccessToken(login.AccessToken)
	require.NoError(t, err)

	user, err := authService.CreateUser("owner@example.com", "password123", auth.RoleUser)
	require.NoError(t, err)
	ownerClaims := &auth.Claims{UserID: user.ID, Email: user.Email, Role: user.Role}

	vol, err := volumeService.Create(ownerClaims, volume.CreateVolumeInput{
		Name:     "Proto",
		DiskPath: storageRoot,
	})
	require.NoError(t, err)

	t.Run("get forbidden for non-owner", func(t *testing.T) {
		otherLogin, err := authService.Login("owner@example.com", "password123")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/volumes/"+vol.ID.String()+"/protocols", nil)
		req.Header.Set("Authorization", "Bearer "+otherLogin.AccessToken)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("admin patch protocols", func(t *testing.T) {
		body, _ := json.Marshal(map[string]any{
			"webdav": map[string]bool{"enabled": true},
		})
		req := httptest.NewRequest(http.MethodPatch, "/api/volumes/"+vol.ID.String()+"/protocols", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+login.AccessToken)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)

		var resp volume.ProtocolsInfo
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		require.True(t, resp.WebDAV.Enabled)
		require.Contains(t, resp.Connection.WebDAVURL, vol.ID.String())
		require.Equal(t, vol.ID.String(), resp.Connection.FTPUsername)
	})

	_ = adminClaims
}
