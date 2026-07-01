package api

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/theo-henon/lcloud/internal/volume"
)

func TestMaskVolume_ClearsDiskAndRootPaths(t *testing.T) {
	vol := volume.Volume{
		ID:       uuid.New(),
		Name:     "Photos",
		DiskPath: "/data/disks/ssd",
		RootPath: "/data/disks/ssd/" + uuid.NewString(),
	}

	masked := maskVolume(vol)
	require.Empty(t, masked.DiskPath)
	require.Empty(t, masked.RootPath)
	require.Equal(t, "Photos", masked.Name)
}
