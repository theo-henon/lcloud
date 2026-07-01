package volume

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestWriteAndReadVolumeConfig(t *testing.T) {
	root := t.TempDir()
	id := uuid.New()
	ownerID := uuid.New()
	created := time.Now().UTC().Truncate(time.Second)

	cfg := VolumeConfig{
		ID:         id,
		Name:       "Photos",
		OwnerID:    ownerID,
		QuotaBytes: 50,
		Filters:    Filters{Mode: FilterModeAllow, Extensions: []string{".jpg"}},
		Encryption: EncryptionConfig{Enabled: false, Method: nil},
		CreatedAt:  created,
		DiskPath:   "/data/disks/ssd",
	}

	require.NoError(t, WriteVolumeConfig(root, cfg))

	read, err := ReadVolumeConfig(root)
	require.NoError(t, err)
	require.Equal(t, cfg.ID, read.ID)
	require.Equal(t, cfg.Name, read.Name)
	require.Equal(t, cfg.Filters, read.Filters)
}

func TestWriteVolumeConfigAtomic(t *testing.T) {
	root := t.TempDir()
	cfg := VolumeConfig{
		ID:        uuid.New(),
		Name:      "Test",
		OwnerID:   uuid.New(),
		CreatedAt: time.Now().UTC(),
		DiskPath:  "/data/disks/ssd",
	}

	require.NoError(t, WriteVolumeConfig(root, cfg))
	_, err := os.Stat(filepath.Join(root, VolumeJSONFile))
	require.NoError(t, err)

	matches, err := filepath.Glob(filepath.Join(root, ".volume.json.*.tmp"))
	require.NoError(t, err)
	require.Empty(t, matches)
}

func TestSyncVolumeFromDisk(t *testing.T) {
	root := t.TempDir()
	id := uuid.New()
	cfg := VolumeConfig{
		ID:        id,
		Name:      "Docs",
		OwnerID:   uuid.New(),
		CreatedAt: time.Now().UTC(),
		DiskPath:  "/data/disks/hdd1",
	}
	require.NoError(t, WriteVolumeConfig(root, cfg))

	vol, err := SyncVolumeFromDisk(root)
	require.NoError(t, err)
	require.Equal(t, id, vol.ID)
	require.Equal(t, root, vol.RootPath)
}
