package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/lcloud?sslmode=disable")
	t.Setenv("JWT_SECRET", "01234567890123456789012345678901")
	t.Setenv("ADMIN_EMAIL", "admin@example.com")
	t.Setenv("ADMIN_PASSWORD", "secret123")
	t.Setenv("STORAGE_BASE_PATH", "./data/storage")
	unsetEnv(t, "JWT_EXPIRY_HOURS")
	unsetEnv(t, "REFRESH_TOKEN_EXPIRY_DAYS")
	unsetEnv(t, "APP_PORT")
	unsetEnv(t, "GIN_MODE")

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, 24, cfg.JWTExpiryHours)
	require.Equal(t, 7, cfg.RefreshTokenExpiryDays)
	require.Equal(t, "8080", cfg.AppPort)
	require.Equal(t, "release", cfg.GinMode)
	require.Equal(t, int64(104857600), cfg.MaxUploadBytes)
}

func TestLoadStorageDiskPaths(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/lcloud?sslmode=disable")
	t.Setenv("JWT_SECRET", "01234567890123456789012345678901")
	t.Setenv("ADMIN_EMAIL", "admin@example.com")
	t.Setenv("ADMIN_PASSWORD", "secret123")
	t.Setenv("STORAGE_BASE_PATH", "./data/storage")
	t.Setenv("STORAGE_DISK_PATHS", "/data/disks/ssd, /data/disks/hdd1")
	t.Setenv("STORAGE_DISK_LABELS", "SSD,HDD 1")

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, []string{"/data/disks/ssd", "/data/disks/hdd1"}, cfg.StorageDiskPaths)
	require.Equal(t, []string{"SSD", "HDD 1"}, cfg.StorageDiskLabels)
}

func TestLoadMissingRequired(t *testing.T) {
	unsetEnv(t, "DATABASE_URL")
	unsetEnv(t, "JWT_SECRET")
	unsetEnv(t, "ADMIN_EMAIL")
	unsetEnv(t, "ADMIN_PASSWORD")
	unsetEnv(t, "STORAGE_BASE_PATH")

	_, err := Load()
	require.Error(t, err)
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	require.NoError(t, os.Unsetenv(key))
}
