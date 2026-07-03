package dashboard

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/theo-henon/lcloud/internal/auth"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&UserDashboardLayout{}))
	return db
}

func TestValidateLayoutValidDefaultUser(t *testing.T) {
	layout := DefaultLayout(auth.RoleUser)
	require.NoError(t, ValidateLayout(layout, auth.RoleUser))
	require.Equal(t, LayoutVersion, layout.Version)
	require.Len(t, layout.Widgets, 6)
}

func TestValidateLayoutValidDefaultAdmin(t *testing.T) {
	layout := DefaultLayout(auth.RoleAdmin)
	require.NoError(t, ValidateLayout(layout, auth.RoleAdmin))
	require.Equal(t, LayoutVersion, layout.Version)
	require.Len(t, layout.Widgets, 6)
}

func TestValidateLayoutUnknownWidget(t *testing.T) {
	layout := Layout{
		Version: LayoutVersion,
		Widgets: []WidgetPlacement{
			{ID: "w1", Type: "not-a-widget", X: 0, Y: 0, W: 4, H: 1},
		},
	}
	err := ValidateLayout(layout, auth.RoleUser)
	require.ErrorIs(t, err, ErrUnknownWidget)
}

func TestValidateLayoutAdminWidgetAsUser(t *testing.T) {
	layout := Layout{
		Version: LayoutVersion,
		Widgets: []WidgetPlacement{
			{ID: "w1", Type: "pending-deletions", X: 0, Y: 0, W: 6, H: 2},
		},
	}
	err := ValidateLayout(layout, auth.RoleUser)
	require.ErrorIs(t, err, ErrWidgetNotAllowed)
}

func TestValidateLayoutOverlap(t *testing.T) {
	layout := Layout{
		Version: LayoutVersion,
		Widgets: []WidgetPlacement{
			{ID: "w1", Type: "welcome", X: 0, Y: 0, W: 4, H: 1},
			{ID: "w2", Type: "quick-actions", X: 2, Y: 0, W: 4, H: 1},
		},
	}
	err := ValidateLayout(layout, auth.RoleUser)
	require.ErrorIs(t, err, ErrInvalidLayout)
}

func TestValidateLayoutOutOfBounds(t *testing.T) {
	layout := Layout{
		Version: LayoutVersion,
		Widgets: []WidgetPlacement{
			{ID: "w1", Type: "welcome", X: 10, Y: 0, W: 4, H: 1},
		},
	}
	err := ValidateLayout(layout, auth.RoleUser)
	require.ErrorIs(t, err, ErrInvalidLayout)
}

func TestValidateLayoutWrongSize(t *testing.T) {
	layout := Layout{
		Version: LayoutVersion,
		Widgets: []WidgetPlacement{
			{ID: "w1", Type: "welcome", X: 0, Y: 0, W: 6, H: 1},
		},
	}
	err := ValidateLayout(layout, auth.RoleUser)
	require.ErrorIs(t, err, ErrInvalidLayout)
}

func TestValidateLayoutValidSimple(t *testing.T) {
	layout := Layout{
		Version: LayoutVersion,
		Widgets: []WidgetPlacement{
			{ID: "w1", Type: "welcome", X: 0, Y: 0, W: 4, H: 1},
			{ID: "w2", Type: "quick-actions", X: 4, Y: 0, W: 4, H: 1},
		},
	}
	require.NoError(t, ValidateLayout(layout, auth.RoleUser))
}

func TestLayoutScanString(t *testing.T) {
	var layout Layout
	raw := `{"version":1,"widgets":[{"id":"w1","type":"welcome","x":0,"y":0,"w":4,"h":1}]}`
	require.NoError(t, layout.Scan(raw))
	require.Equal(t, LayoutVersion, layout.Version)
	require.Len(t, layout.Widgets, 1)
	require.Equal(t, "welcome", layout.Widgets[0].Type)
}

func TestGetLayoutNoRowReturnsDefault(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	userID := uuid.New()
	claims := &auth.Claims{UserID: userID, Role: auth.RoleUser}
	result, err := service.GetLayout(t.Context(), claims)
	require.NoError(t, err)
	require.Nil(t, result.UpdatedAt)
	require.Equal(t, DefaultLayout(auth.RoleUser), result.Layout)
	require.NoError(t, ValidateLayout(result.Layout, auth.RoleUser))
	require.Len(t, result.Catalog, 6)
}

func TestGetLayoutAdminCatalog(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	claims := &auth.Claims{UserID: uuid.New(), Role: auth.RoleAdmin}
	result, err := service.GetLayout(t.Context(), claims)
	require.NoError(t, err)
	require.Len(t, result.Catalog, 7)
	require.NoError(t, ValidateLayout(result.Layout, auth.RoleAdmin))
}

func TestSaveLayoutDefaultAdmin(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	claims := &auth.Claims{UserID: uuid.New(), Role: auth.RoleAdmin}
	layout := DefaultLayout(auth.RoleAdmin)
	_, err := service.SaveLayout(t.Context(), claims, layout)
	require.NoError(t, err)
}

func TestSaveLayoutRoundTrip(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	userID := uuid.New()
	claims := &auth.Claims{UserID: userID, Role: auth.RoleUser}
	layout := Layout{
		Version: LayoutVersion,
		Widgets: []WidgetPlacement{
			{ID: "custom-welcome", Type: "welcome", X: 0, Y: 0, W: 4, H: 1},
		},
	}

	saved, err := service.SaveLayout(t.Context(), claims, layout)
	require.NoError(t, err)
	require.NotNil(t, saved.UpdatedAt)

	fetched, err := service.GetLayout(t.Context(), claims)
	require.NoError(t, err)
	require.NotNil(t, fetched.UpdatedAt)
	require.Equal(t, layout, fetched.Layout)
}

func TestResetLayoutDeletesRow(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	userID := uuid.New()
	claims := &auth.Claims{UserID: userID, Role: auth.RoleUser}
	layout := Layout{
		Version: LayoutVersion,
		Widgets: []WidgetPlacement{
			{ID: "w1", Type: "welcome", X: 0, Y: 0, W: 4, H: 1},
		},
	}
	_, err := service.SaveLayout(t.Context(), claims, layout)
	require.NoError(t, err)

	require.NoError(t, service.ResetLayout(t.Context(), userID))

	result, err := service.GetLayout(t.Context(), claims)
	require.NoError(t, err)
	require.Nil(t, result.UpdatedAt)
	require.Equal(t, DefaultLayout(auth.RoleUser), result.Layout)
}

func TestSaveLayoutRejectsInvalid(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	claims := &auth.Claims{UserID: uuid.New(), Role: auth.RoleUser}
	_, err := service.SaveLayout(t.Context(), claims, Layout{Version: 2, Widgets: []WidgetPlacement{
		{ID: "w1", Type: "welcome", X: 0, Y: 0, W: 4, H: 1},
	}})
	require.ErrorIs(t, err, ErrInvalidLayout)
}
