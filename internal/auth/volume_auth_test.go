package auth

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func insertTestVolume(t *testing.T, db *gorm.DB, volumeID, ownerID uuid.UUID) {
	t.Helper()
	require.NoError(t, db.Exec(`
		CREATE TABLE IF NOT EXISTS volumes (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			owner_id TEXT NOT NULL,
			disk_path TEXT NOT NULL,
			root_path TEXT NOT NULL,
			quota_bytes INTEGER NOT NULL DEFAULT 0,
			used_bytes INTEGER NOT NULL DEFAULT 0,
			filters TEXT NOT NULL DEFAULT '{}',
			protocols TEXT NOT NULL DEFAULT '{}',
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error)
	require.NoError(t, db.Exec(
		`INSERT INTO volumes (id, name, owner_id, disk_path, root_path, quota_bytes, used_bytes, filters, protocols, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, 0, 0, '{}', '{}', datetime('now'), datetime('now'))`,
		volumeID.String(), "test", ownerID.String(), "/data", "/data/"+volumeID.String(),
	).Error)
}

func TestAuthenticateForVolumeOwner(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(t, db)

	owner, err := service.CreateUser("owner@example.com", "password123", RoleUser)
	require.NoError(t, err)

	volumeID := uuid.New()
	insertTestVolume(t, db, volumeID, owner.ID)

	claims, err := service.AuthenticateForVolume(volumeID, "password123")
	require.NoError(t, err)
	require.Equal(t, owner.ID, claims.UserID)
	require.Equal(t, "owner@example.com", claims.Email)
}

func TestAuthenticateForVolumeAdmin(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(t, db)

	require.NoError(t, service.SeedAdmin("admin@example.com", "adminpass1"))
	owner, err := service.CreateUser("owner@example.com", "password123", RoleUser)
	require.NoError(t, err)

	volumeID := uuid.New()
	insertTestVolume(t, db, volumeID, owner.ID)

	claims, err := service.AuthenticateForVolume(volumeID, "adminpass1")
	require.NoError(t, err)
	require.Equal(t, RoleAdmin, claims.Role)
}

func TestAuthenticateForVolumeWrongPassword(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(t, db)

	owner, err := service.CreateUser("owner@example.com", "password123", RoleUser)
	require.NoError(t, err)

	volumeID := uuid.New()
	insertTestVolume(t, db, volumeID, owner.ID)

	_, err = service.AuthenticateForVolume(volumeID, "wrong-password")
	require.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestValidateCredentials(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(t, db)

	_, err := service.CreateUser("user@example.com", "password123", RoleUser)
	require.NoError(t, err)

	claims, err := service.ValidateCredentials("user@example.com", "password123")
	require.NoError(t, err)
	require.Equal(t, "user@example.com", claims.Email)

	_, err = service.ValidateCredentials("user@example.com", "bad")
	require.ErrorIs(t, err, ErrInvalidCredentials)
}
