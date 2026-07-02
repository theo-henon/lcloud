package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/theo-henon/lcloud/internal/api"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/config"
	"github.com/theo-henon/lcloud/internal/indexer"
	"github.com/theo-henon/lcloud/internal/monitoring"
	"github.com/theo-henon/lcloud/internal/plugin"
	"github.com/theo-henon/lcloud/internal/settings"
	"github.com/theo-henon/lcloud/internal/volume"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:embed all:static
var staticEmbed embed.FS

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	if err := db.AutoMigrate(
		&auth.User{},
		&auth.RefreshToken{},
		&volume.Volume{},
		&settings.InstanceSettings{},
		&plugin.Plugin{},
		&plugin.PluginLogEntry{},
	); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	authService := auth.NewService(db, cfg.JWTSecret, cfg.JWTExpiryHours, cfg.RefreshTokenExpiryDays)
	if err := authService.SeedAdmin(cfg.AdminEmail, cfg.AdminPassword); err != nil {
		log.Fatalf("seed admin: %v", err)
	}

	settingsService := settings.NewService(db)
	if err := settingsService.EnsureDefaults(); err != nil {
		log.Fatalf("settings defaults: %v", err)
	}

	diskRegistry := volume.NewDiskRegistry(cfg)
	indexManager := indexer.NewIndexManager()
	volumeService := volume.NewService(db, diskRegistry, indexManager)
	monitoringService := monitoring.NewService(volumeService, diskRegistry, settingsService)
	fileService := volume.NewFileService(volumeService, indexManager, cfg.MaxUploadBytes, monitoringService.StatsCache())

	pluginService := plugin.NewService(db, cfg.PluginsPath, volumeService, nil)
	volumeService.SetEventPublisher(pluginService.Publisher())
	fileService.SetEventPublisher(pluginService.Publisher())

	ctx := context.Background()
	if err := pluginService.Start(ctx); err != nil {
		log.Fatalf("plugin startup: %v", err)
	}
	defer pluginService.Stop()

	staticFS, err := fs.Sub(staticEmbed, "static")
	if err != nil {
		log.Fatalf("static fs: %v", err)
	}

	router := api.NewRouter(api.RouterConfig{
		AuthService:       authService,
		DiskRegistry:      diskRegistry,
		VolumeService:     volumeService,
		FileService:       fileService,
		MonitoringService: monitoringService,
		SettingsService:   settingsService,
		PluginService:     pluginService,
		IndexManager:      indexManager,
		MaxUploadBytes:    cfg.MaxUploadBytes,
		StaticFS:          staticFS,
		GinMode:           cfg.GinMode,
	})

	addr := ":" + cfg.AppPort
	server := &http.Server{Addr: addr, Handler: router}

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		pluginService.Stop()
		_ = server.Close()
	}()

	log.Printf("lcloud listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}
