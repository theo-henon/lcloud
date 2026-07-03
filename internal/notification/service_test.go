package notification

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/settings"
	"github.com/theo-henon/lcloud/internal/task"
	"github.com/theo-henon/lcloud/internal/volume"
	"gorm.io/gorm"
)

func setupNotificationTest(t *testing.T) (*Service, *auth.Service, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&auth.User{},
		&Notification{},
		&settings.InstanceSettings{},
		&volume.Volume{},
		&task.Task{},
	))

	authService := auth.NewService(db, "01234567890123456789012345678901", 24, 7)
	require.NoError(t, authService.SeedAdmin("admin@example.com", "adminpass1"))

	settingsService := settings.NewService(db)
	require.NoError(t, settingsService.EnsureDefaults())

	diskRegistry := volume.NewDiskRegistry(nil)
	volumeService := volume.NewService(db, diskRegistry, nil)
	taskService := task.NewService(db, volumeService, authService, nil, nil)

	service := NewService(db, authService, volumeService, taskService, settingsService)
	return service, authService, db
}

func TestListScopedToUser(t *testing.T) {
	service, authService, db := setupNotificationTest(t)

	var adminUser auth.User
	require.NoError(t, db.Where("email = ?", "admin@example.com").First(&adminUser).Error)

	otherUser, err := authService.CreateUser("user@example.com", "userpass12", auth.RoleUser)
	require.NoError(t, err)

	ctx := context.Background()
	source := "evt-1"
	require.NoError(t, service.create(ctx, createInput{
		UserID:        adminUser.ID,
		Type:          TypeVolumeUsageAlert,
		Title:         "Admin alert",
		Body:          "body",
		LinkPath:      "/volumes",
		SourceEventID: &source,
	}))
	source2 := "evt-2"
	require.NoError(t, service.create(ctx, createInput{
		UserID:        otherUser.ID,
		Type:          TypeVolumeUsageAlert,
		Title:         "User alert",
		Body:          "body",
		LinkPath:      "/volumes",
		SourceEventID: &source2,
	}))

	items, total, err := service.List(ctx, adminUser.ID, 30, 0)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	require.Equal(t, "Admin alert", items[0].Title)
}

func TestDismissHidesFromList(t *testing.T) {
	service, authService, db := setupNotificationTest(t)

	var adminUser auth.User
	require.NoError(t, db.Where("email = ?", "admin@example.com").First(&adminUser).Error)

	ctx := context.Background()
	source := "evt-dismiss"
	require.NoError(t, service.create(ctx, createInput{
		UserID:        adminUser.ID,
		Type:          TypeTaskFailed,
		Title:         "Failed",
		Body:          "error",
		LinkPath:      "/tasks",
		SourceEventID: &source,
	}))

	items, _, err := service.List(ctx, adminUser.ID, 30, 0)
	require.NoError(t, err)
	require.Len(t, items, 1)

	count, err := service.UnreadCount(ctx, adminUser.ID)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)

	require.NoError(t, service.Dismiss(ctx, adminUser.ID, items[0].ID))

	items, total, err := service.List(ctx, adminUser.ID, 30, 0)
	require.NoError(t, err)
	require.Equal(t, int64(0), total)
	require.Empty(t, items)

	count, err = service.UnreadCount(ctx, adminUser.ID)
	require.NoError(t, err)
	require.Equal(t, int64(0), count)

	_ = authService
}

func TestCreateIdempotentOnSourceEventID(t *testing.T) {
	service, _, db := setupNotificationTest(t)

	var adminUser auth.User
	require.NoError(t, db.Where("email = ?", "admin@example.com").First(&adminUser).Error)

	ctx := context.Background()
	source := "same-event"
	input := createInput{
		UserID:        adminUser.ID,
		Type:          TypeVolumeUsageAlert,
		Title:         "Alert",
		Body:          "body",
		LinkPath:      "/volumes",
		SourceEventID: &source,
	}
	require.NoError(t, service.create(ctx, input))
	require.NoError(t, service.create(ctx, input))

	items, total, err := service.List(ctx, adminUser.ID, 30, 0)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, items, 1)
}

func TestMarkReadAndMarkAllRead(t *testing.T) {
	service, _, db := setupNotificationTest(t)

	var adminUser auth.User
	require.NoError(t, db.Where("email = ?", "admin@example.com").First(&adminUser).Error)

	ctx := context.Background()
	for i, src := range []string{"a", "b"} {
		source := src
		require.NoError(t, service.create(ctx, createInput{
			UserID:        adminUser.ID,
			Type:          TypeVolumeUsageAlert,
			Title:         "Alert",
			Body:          "body",
			LinkPath:      "/volumes",
			SourceEventID: &source,
		}))
		_ = i
	}

	items, _, err := service.List(ctx, adminUser.ID, 30, 0)
	require.NoError(t, err)
	require.Len(t, items, 2)

	_, err = service.MarkRead(ctx, adminUser.ID, items[0].ID)
	require.NoError(t, err)

	count, err := service.UnreadCount(ctx, adminUser.ID)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)

	updated, err := service.MarkAllRead(ctx, adminUser.ID)
	require.NoError(t, err)
	require.Equal(t, int64(1), updated)

	count, err = service.UnreadCount(ctx, adminUser.ID)
	require.NoError(t, err)
	require.Equal(t, int64(0), count)
}

func TestRetentionCap(t *testing.T) {
	service, _, db := setupNotificationTest(t)

	var adminUser auth.User
	require.NoError(t, db.Where("email = ?", "admin@example.com").First(&adminUser).Error)

	ctx := context.Background()
	for i := 0; i < maxNotificationsPerUser+5; i++ {
		source := uuid.NewString()
		require.NoError(t, service.create(ctx, createInput{
			UserID:        adminUser.ID,
			Type:          TypeVolumeUsageAlert,
			Title:         "Alert",
			Body:          "body",
			LinkPath:      "/volumes",
			SourceEventID: &source,
		}))
	}

	var count int64
	require.NoError(t, db.Model(&Notification{}).Where("user_id = ?", adminUser.ID).Count(&count).Error)
	require.Equal(t, int64(maxNotificationsPerUser), count)
}

func TestDismissNotFoundForOtherUser(t *testing.T) {
	service, authService, db := setupNotificationTest(t)

	var adminUser auth.User
	require.NoError(t, db.Where("email = ?", "admin@example.com").First(&adminUser).Error)

	otherUser, err := authService.CreateUser("other@example.com", "otherpass1", auth.RoleUser)
	require.NoError(t, err)

	ctx := context.Background()
	source := "evt-x"
	require.NoError(t, service.create(ctx, createInput{
		UserID:        adminUser.ID,
		Type:          TypeVolumeUsageAlert,
		Title:         "Alert",
		Body:          "body",
		LinkPath:      "/volumes",
		SourceEventID: &source,
	}))

	items, _, err := service.List(ctx, adminUser.ID, 30, 0)
	require.NoError(t, err)

	err = service.Dismiss(ctx, otherUser.ID, items[0].ID)
	require.ErrorIs(t, err, ErrNotFound)
}
