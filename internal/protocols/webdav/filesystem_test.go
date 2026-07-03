package webdav

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/config"
	"github.com/theo-henon/lcloud/internal/indexer"
	"github.com/theo-henon/lcloud/internal/protocols"
	"github.com/theo-henon/lcloud/internal/settings"
	"github.com/theo-henon/lcloud/internal/volume"
	"github.com/glebarez/sqlite"
	gowebdav "golang.org/x/net/webdav"
	"gorm.io/gorm"
)

func TestPROPFINDRoot(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&auth.User{}, &volume.Volume{}, &settings.InstanceSettings{}))

	authService := auth.NewService(db, "01234567890123456789012345678901", 24, 7)
	_, err = authService.CreateUser("admin@example.com", "adminpass1", auth.RoleAdmin)
	require.NoError(t, err)

	settingsService := settings.NewService(db)
	require.NoError(t, settingsService.EnsureDefaults())
	_, err = settingsService.Update(settings.UpdateSettingsInput{
		ProtocolsWebDAVEnabled: boolPtr(true),
	})
	require.NoError(t, err)

	storageRoot := t.TempDir()
	cfg := &config.Config{
		StorageBasePath:  storageRoot,
		StorageDiskPaths: []string{storageRoot},
		MaxUploadBytes:   10 * 1024 * 1024,
	}
	indexManager := indexer.NewIndexManager()
	volumeService := volume.NewService(db, volume.NewDiskRegistry(cfg), indexManager)
	fileService := volume.NewFileService(volumeService, indexManager, cfg.MaxUploadBytes, nil)

	admin, err := authService.ValidateCredentials("admin@example.com", "adminpass1")
	require.NoError(t, err)

	vol, err := volumeService.Create(admin, volume.CreateVolumeInput{Name: "WebDAV", DiskPath: storageRoot})
	require.NoError(t, err)

	enabled := true
	_, err = volumeService.UpdateProtocols(admin, vol.ID, volume.UpdateProtocolsInput{
		WebDAV: &volume.ProtocolToggle{Enabled: enabled},
	})
	require.NoError(t, err)

	fs := newVolumeFS(fileService, admin, vol.ID)
	dir, err := fs.OpenFile(t.Context(), "/", os.O_RDONLY, 0)
	require.NoError(t, err)
	entries, err := dir.Readdir(-1)
	require.NoError(t, err)
	t.Logf("root entries: %d", len(entries))

	prefix := "/dav/volumes/" + vol.ID.String() + "/"
	direct := &gowebdav.Handler{Prefix: prefix, FileSystem: fs, LockSystem: gowebdav.NewMemLS()}
	directReq := httptest.NewRequest("PROPFIND", prefix, nil)
	directReq.Header.Set("Depth", "1")
	directRec := httptest.NewRecorder()
	direct.ServeHTTP(directRec, directReq)
	t.Logf("direct webdav status=%d", directRec.Code)

	gateway := protocols.NewGateway(authService, volumeService, fileService, settingsService, 2121)

	router := gin.New()
	RegisterRoutes(router, gateway)

	url := "/dav/volumes/" + vol.ID.String() + "/"
	req := httptest.NewRequest("PROPFIND", url, nil)
	req.SetBasicAuth("admin@example.com", "adminpass1")
	req.Header.Set("Depth", "1")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	t.Logf("status=%d body=%q", rec.Code, rec.Body.String())
	require.Equal(t, http.StatusMultiStatus, rec.Code)
}

func boolPtr(v bool) *bool { return &v }
