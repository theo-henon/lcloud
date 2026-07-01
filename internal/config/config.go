package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL            string
	JWTSecret              string
	JWTExpiryHours         int
	RefreshTokenExpiryDays int
	AdminEmail             string
	AdminPassword          string
	StorageBasePath        string
	AppPort                string
	GinMode                string
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		JWTSecret:              os.Getenv("JWT_SECRET"),
		AdminEmail:             os.Getenv("ADMIN_EMAIL"),
		AdminPassword:          os.Getenv("ADMIN_PASSWORD"),
		StorageBasePath:        os.Getenv("STORAGE_BASE_PATH"),
		AppPort:                getEnv("APP_PORT", "8080"),
		GinMode:                getEnv("GIN_MODE", "release"),
		JWTExpiryHours:         getEnvInt("JWT_EXPIRY_HOURS", 24),
		RefreshTokenExpiryDays: getEnvInt("REFRESH_TOKEN_EXPIRY_DAYS", 7),
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

	return cfg, nil
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
