package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "file:" + t.Name() + "?mode=memory&cache=private"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&User{}, &RefreshToken{}))
	return db
}

func newTestService(t *testing.T, db *gorm.DB) *Service {
	t.Helper()
	return NewService(db, "01234567890123456789012345678901", 24, 7)
}

func TestLoginRefreshLogout(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(t, db)

	_, err := service.CreateUser("user@example.com", "password123", RoleUser)
	require.NoError(t, err)

	login, err := service.Login("user@example.com", "password123")
	require.NoError(t, err)
	require.NotEmpty(t, login.AccessToken)
	require.NotEmpty(t, login.RefreshToken)

	claims, err := service.ValidateAccessToken(login.AccessToken)
	require.NoError(t, err)
	require.Equal(t, "user@example.com", claims.Email)

	refreshed, err := service.Refresh(login.RefreshToken)
	require.NoError(t, err)
	require.NotEmpty(t, refreshed.AccessToken)
	require.NotEqual(t, login.RefreshToken, refreshed.RefreshToken)

	err = service.Logout(refreshed.RefreshToken)
	require.NoError(t, err)

	_, err = service.Refresh(refreshed.RefreshToken)
	require.ErrorIs(t, err, ErrRevokedToken)
}

func TestLoginInvalidCredentials(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(t, db)

	_, err := service.CreateUser("user@example.com", "password123", RoleUser)
	require.NoError(t, err)

	_, err = service.Login("user@example.com", "wrong-password")
	require.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestSeedAdminOnlyOnce(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(t, db)

	require.NoError(t, service.SeedAdmin("admin@example.com", "adminpass1"))
	require.NoError(t, service.SeedAdmin("other@example.com", "otherpass1"))

	var count int64
	require.NoError(t, db.Model(&User{}).Count(&count).Error)
	require.Equal(t, int64(1), count)
}

func TestExpiredRefreshToken(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(t, db)

	user, err := service.CreateUser("user@example.com", "password123", RoleUser)
	require.NoError(t, err)

	raw, err := service.issueRefreshToken(user)
	require.NoError(t, err)

	require.NoError(t, db.Model(&RefreshToken{}).
		Where("user_id = ?", user.ID).
		Update("expires_at", time.Now().Add(-time.Hour)).Error)

	_, err = service.Refresh(raw)
	require.ErrorIs(t, err, ErrExpiredToken)
}

func TestCreateUserInvalidRole(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(t, db)

	_, err := service.CreateUser("user@example.com", "password123", Role("superadmin"))
	require.ErrorIs(t, err, ErrInvalidRole)
}

func TestCreateUserEmailTaken(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(t, db)

	_, err := service.CreateUser("user@example.com", "password123", RoleUser)
	require.NoError(t, err)

	_, err = service.CreateUser("user@example.com", "anotherpass", RoleUser)
	require.ErrorIs(t, err, ErrEmailTaken)
}

func TestValidateAccessTokenInvalid(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(t, db)

	_, err := service.ValidateAccessToken("not-a-token")
	require.ErrorIs(t, err, ErrInvalidToken)
}

func TestGetUserByIDNotFound(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(t, db)

	_, err := service.GetUserByID(uuid.New())
	require.ErrorIs(t, err, ErrInvalidToken)
}

func TestRefreshWithInvalidToken(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(t, db)

	_, err := service.Refresh("invalid-token")
	require.ErrorIs(t, err, ErrInvalidToken)
}
