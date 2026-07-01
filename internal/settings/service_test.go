package settings

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/theo-henon/lcloud/internal/auth"
	"gorm.io/gorm"
)

func setupSettingsTest(t *testing.T) *Service {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&InstanceSettings{}))
	service := NewService(db)
	require.NoError(t, service.EnsureDefaults())
	return service
}

func TestShouldMaskFor_WhenEnabled(t *testing.T) {
	service := setupSettingsTest(t)
	updated, err := service.Update(UpdateSettingsInput{MaskDiskNames: ptrBool(true)})
	require.NoError(t, err)
	require.True(t, updated.MaskDiskNames)

	mask, err := service.ShouldMaskFor(&auth.Claims{Role: auth.RoleAdmin})
	require.NoError(t, err)
	require.True(t, mask)
}

func TestShouldMaskFor_RegularUser(t *testing.T) {
	service := setupSettingsTest(t)
	_, err := service.Update(UpdateSettingsInput{MaskDiskNames: ptrBool(true)})
	require.NoError(t, err)

	mask, err := service.ShouldMaskFor(&auth.Claims{UserID: uuid.New(), Role: auth.RoleUser})
	require.NoError(t, err)
	require.True(t, mask)
}

func ptrBool(v bool) *bool {
	return &v
}
