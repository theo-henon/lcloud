package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const defaultMaxUploadBytes = 100 * 1024 * 1024 // 100MB

type Config struct {
	DatabaseURL            string
	JWTSecret              string
	JWTExpiryHours         int
	RefreshTokenExpiryDays int
	AdminEmail             string
	AdminPassword          string
	StorageBasePath        string
	StorageDiskPaths       []string
	StorageDiskLabels      []string
	MaxUploadBytes         int64
	AppPort                string
	GinMode                string
	PluginsPath            string
	FTPPort                int
	FTPPasvMin             int
	FTPPasvMax             int
	FTPPasvAddress         string
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		JWTSecret:              os.Getenv("JWT_SECRET"),
		AdminEmail:             os.Getenv("ADMIN_EMAIL"),
		AdminPassword:          os.Getenv("ADMIN_PASSWORD"),
		StorageBasePath:        os.Getenv("STORAGE_BASE_PATH"),
		StorageDiskPaths:       parseCSV(os.Getenv("STORAGE_DISK_PATHS")),
		StorageDiskLabels:      parseCSV(os.Getenv("STORAGE_DISK_LABELS")),
		AppPort:                getEnv("APP_PORT", "8080"),
		GinMode:                getEnv("GIN_MODE", "release"),
		JWTExpiryHours:         getEnvInt("JWT_EXPIRY_HOURS", 24),
		RefreshTokenExpiryDays: getEnvInt("REFRESH_TOKEN_EXPIRY_DAYS", 7),
		MaxUploadBytes:         getEnvInt64("MAX_UPLOAD_BYTES", defaultMaxUploadBytes),
		PluginsPath:            getEnv("PLUGINS_PATH", "./plugins"),
		FTPPort:                getEnvInt("FTP_PORT", 2121),
		FTPPasvMin:             getEnvInt("FTP_PASV_MIN", 30000),
		FTPPasvMax:             getEnvInt("FTP_PASV_MAX", 30010),
		FTPPasvAddress:         getEnv("FTP_PASV_ADDRESS", "127.0.0.1"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if cfg.AdminEmail == "" {
		return nil, fmt.Errorf("ADMIN_EMAIL is required")
	}
	if cfg.AdminPassword == "" {
		return nil, fmt.Errorf("ADMIN_PASSWORD is required")
	}
	if len(cfg.AdminPassword) < 8 {
		return nil, fmt.Errorf("ADMIN_PASSWORD must be at least 8 characters")
	}
	if cfg.StorageBasePath == "" {
		return nil, fmt.Errorf("STORAGE_BASE_PATH is required")
	}
	if cfg.MaxUploadBytes <= 0 {
		return nil, fmt.Errorf("MAX_UPLOAD_BYTES must be positive")
	}

	return cfg, nil
}

func parseCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func getEnvInt64(key string, fallback int64) int64 {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return fallback
	}
	return value
}
