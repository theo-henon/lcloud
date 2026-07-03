package volume

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/theo-henon/lcloud/internal/auth"
)

func TestUpdateProtocolsWritesDiskFirst(t *testing.T) {
	service, _, claims, _ := setupVolumeTest(t)

	vol, err := service.Create(claims, CreateVolumeInput{
		Name:     "Protocols",
		DiskPath: service.disks.ResolvedPaths()[0],
	})
	require.NoError(t, err)
	require.False(t, vol.Protocols.WebDAV.Enabled)
	require.False(t, vol.Protocols.FTP.Enabled)

	enabled := true
	info, err := service.UpdateProtocols(claims, vol.ID, UpdateProtocolsInput{
		WebDAV: &ProtocolToggle{Enabled: enabled},
	})
	require.NoError(t, err)
	require.True(t, info.WebDAV.Enabled)

	cfg, err := ReadVolumeConfig(vol.RootPath)
	require.NoError(t, err)
	require.True(t, cfg.Protocols.WebDAV.Enabled)

	reloaded, err := service.GetByID(vol.ID)
	require.NoError(t, err)
	require.True(t, reloaded.Protocols.WebDAV.Enabled)
}

func TestProtocolEnabledGate(t *testing.T) {
	service, _, _, _ := setupVolumeTest(t)

	vol := &Volume{
		ID: uuid.New(),
		Protocols: ProtocolsConfig{
			WebDAV: ProtocolToggle{Enabled: true},
			FTP:    ProtocolToggle{Enabled: true},
		},
	}

	require.True(t, service.ProtocolEnabled(vol, ProtocolWebDAV, true, false))
	require.False(t, service.ProtocolEnabled(vol, ProtocolWebDAV, false, true))

	vol.Protocols.WebDAV.Enabled = false
	require.False(t, service.ProtocolEnabled(vol, ProtocolWebDAV, true, true))
}

func TestGetProtocolsForbidden(t *testing.T) {
	service, _, ownerClaims, _ := setupVolumeTest(t)
	otherClaims := &auth.Claims{UserID: uuid.New(), Role: auth.RoleUser}

	vol, err := service.Create(ownerClaims, CreateVolumeInput{
		Name:     "Private",
		DiskPath: service.disks.ResolvedPaths()[0],
	})
	require.NoError(t, err)

	_, err = service.GetProtocols(otherClaims, vol.ID, "localhost", 2121)
	require.ErrorIs(t, err, ErrForbidden)
}
