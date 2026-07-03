package protocols_test

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/config"
	"github.com/theo-henon/lcloud/internal/indexer"
	"github.com/theo-henon/lcloud/internal/protocols"
	"github.com/theo-henon/lcloud/internal/settings"
	"github.com/theo-henon/lcloud/internal/volume"
	"gorm.io/gorm"
)

func setupGatewayTest(t *testing.T) (*protocols.Gateway, *auth.Service, *volume.Service, *auth.Claims, uuid.UUID) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&auth.User{}, &volume.Volume{}, &settings.InstanceSettings{}))

	authService := auth.NewService(db, "01234567890123456789012345678901", 24, 7)
	user, err := authService.CreateUser("user@example.com", "password123", auth.RoleUser)
	require.NoError(t, err)

	settingsService := settings.NewService(db)
	require.NoError(t, settingsService.EnsureDefaults())
	_, err = settingsService.Update(settings.UpdateSettingsInput{
		ProtocolsFTPEnabled: ptrBool(true),
	})
	require.NoError(t, err)

	storageRoot := t.TempDir()
	cfg := &config.Config{
		StorageBasePath:  storageRoot,
		StorageDiskPaths: []string{storageRoot},
		MaxUploadBytes:   10 * 1024 * 1024,
	}
	indexManager := indexer.NewIndexManager()
	volumeService := volume.NewService(db, volume.NewDiskRegistry(cfg), indexManager)
	fileService := volume.NewFileService(volumeService, indexManager, cfg.MaxUploadBytes, nil)

	claims := &auth.Claims{UserID: user.ID, Email: user.Email, Role: user.Role}
	vol, err := volumeService.Create(claims, volume.CreateVolumeInput{
		Name:     "FTP",
		DiskPath: storageRoot,
	})
	require.NoError(t, err)

	enabled := true
	_, err = volumeService.UpdateProtocols(claims, vol.ID, volume.UpdateProtocolsInput{
		FTP: &volume.ProtocolToggle{Enabled: enabled},
	})
	require.NoError(t, err)

	gateway := protocols.NewGateway(authService, volumeService, fileService, settingsService, 2121)
	return gateway, authService, volumeService, claims, vol.ID
}

func ptrBool(v bool) *bool { return &v }

func TestFTPLoginSuccess(t *testing.T) {
	gateway, _, _, _, volumeID := setupGatewayTest(t)

	claims, parsedID, err := gateway.FTPLogin("127.0.0.1:1234", volumeID.String(), "password123")
	require.NoError(t, err)
	require.Equal(t, volumeID, parsedID)
	require.Equal(t, "user@example.com", claims.Email)
}

func TestFTPLoginDisabledProtocol(t *testing.T) {
	gateway, _, volumeService, claims, volumeID := setupGatewayTest(t)

	disabled := false
	_, err := volumeService.UpdateProtocols(claims, volumeID, volume.UpdateProtocolsInput{
		FTP: &volume.ProtocolToggle{Enabled: disabled},
	})
	require.NoError(t, err)

	_, _, err = gateway.FTPLogin("127.0.0.1:1234", volumeID.String(), "password123")
	require.ErrorIs(t, err, volume.ErrProtocolsDisabled)
}

func TestFTPLoginInvalidUUID(t *testing.T) {
	gateway, _, _, _, _ := setupGatewayTest(t)

	_, _, err := gateway.FTPLogin("127.0.0.1:1234", "not-a-uuid", "password123")
	require.ErrorIs(t, err, auth.ErrInvalidCredentials)
}
