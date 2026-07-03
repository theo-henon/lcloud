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
	require.ErrorIs(t, err, ErrUserNotFound)
}

func TestListUsers(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(t, db)

	require.NoError(t, service.SeedAdmin("admin@example.com", "adminpass1"))
	_, err := service.CreateUser("member@example.com", "password123", RoleUser)
	require.NoError(t, err)

	users, err := service.ListUsers()
	require.NoError(t, err)
	require.Len(t, users, 2)
	require.Equal(t, "admin@example.com", users[0].Email)
	require.Equal(t, "member@example.com", users[1].Email)
}

func TestLoginDisabledUser(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(t, db)

	user, err := service.CreateUser("user@example.com", "password123", RoleUser)
	require.NoError(t, err)

	_, err = service.PatchUser(user.ID, PatchUserInput{Disabled: ptrBool(true)})
	require.NoError(t, err)

	_, err = service.Login("user@example.com", "password123")
	require.ErrorIs(t, err, ErrUserDisabled)
}

func TestRefreshDisabledUser(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(t, db)

	user, err := service.CreateUser("user@example.com", "password123", RoleUser)
	require.NoError(t, err)

	login, err := service.Login("user@example.com", "password123")
	require.NoError(t, err)

	_, err = service.PatchUser(user.ID, PatchUserInput{Disabled: ptrBool(true)})
	require.NoError(t, err)

	_, err = service.Refresh(login.RefreshToken)
	require.ErrorIs(t, err, ErrRevokedToken)
}

func TestPatchUserDisableRevokesRefreshTokens(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(t, db)

	user, err := service.CreateUser("user@example.com", "password123", RoleUser)
	require.NoError(t, err)

	login, err := service.Login("user@example.com", "password123")
	require.NoError(t, err)

	_, err = service.PatchUser(user.ID, PatchUserInput{Disabled: ptrBool(true)})
	require.NoError(t, err)

	_, err = service.Refresh(login.RefreshToken)
	require.ErrorIs(t, err, ErrRevokedToken)
}

func TestPatchUserEnable(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(t, db)

	user, err := service.CreateUser("user@example.com", "password123", RoleUser)
	require.NoError(t, err)

	_, err = service.PatchUser(user.ID, PatchUserInput{Disabled: ptrBool(true)})
	require.NoError(t, err)

	updated, err := service.PatchUser(user.ID, PatchUserInput{Disabled: ptrBool(false)})
	require.NoError(t, err)
	require.Nil(t, updated.DisabledAt)

	_, err = service.Login("user@example.com", "password123")
	require.NoError(t, err)
}

func TestPatchUserLastAdminCannotDisable(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(t, db)

	require.NoError(t, service.SeedAdmin("admin@example.com", "adminpass1"))

	var admin User
	require.NoError(t, db.Where("email = ?", "admin@example.com").First(&admin).Error)

	_, err := service.PatchUser(admin.ID, PatchUserInput{Disabled: ptrBool(true)})
	require.ErrorIs(t, err, ErrLastAdmin)
}

func TestPatchUserLastAdminCannotDemote(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(t, db)

	require.NoError(t, service.SeedAdmin("admin@example.com", "adminpass1"))

	var admin User
	require.NoError(t, db.Where("email = ?", "admin@example.com").First(&admin).Error)

	role := RoleUser
	_, err := service.PatchUser(admin.ID, PatchUserInput{Role: &role})
	require.ErrorIs(t, err, ErrLastAdmin)
}

func TestPatchUserPromoteToAdmin(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(t, db)

	user, err := service.CreateUser("user@example.com", "password123", RoleUser)
	require.NoError(t, err)

	role := RoleAdmin
	updated, err := service.PatchUser(user.ID, PatchUserInput{Role: &role})
	require.NoError(t, err)
	require.Equal(t, RoleAdmin, updated.Role)
}

func TestPatchUserDemoteWithTwoAdmins(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(t, db)

	require.NoError(t, service.SeedAdmin("admin@example.com", "adminpass1"))
	other, err := service.CreateUser("other@example.com", "password123", RoleAdmin)
	require.NoError(t, err)

	role := RoleUser
	updated, err := service.PatchUser(other.ID, PatchUserInput{Role: &role})
	require.NoError(t, err)
	require.Equal(t, RoleUser, updated.Role)
}

func TestPatchUserResetPassword(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(t, db)

	_, err := service.CreateUser("user@example.com", "password123", RoleUser)
	require.NoError(t, err)

	password := "newpassword1"
	var user User
	require.NoError(t, db.Where("email = ?", "user@example.com").First(&user).Error)

	_, err = service.PatchUser(user.ID, PatchUserInput{Password: &password})
	require.NoError(t, err)

	_, err = service.Login("user@example.com", "password123")
	require.ErrorIs(t, err, ErrInvalidCredentials)

	_, err = service.Login("user@example.com", "newpassword1")
	require.NoError(t, err)
}

func ptrBool(v bool) *bool {
	return &v
}

func TestRefreshWithInvalidToken(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(t, db)

	_, err := service.Refresh("invalid-token")
	require.ErrorIs(t, err, ErrInvalidToken)
}
