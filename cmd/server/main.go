package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/theo-henon/lcloud/internal/api"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/config"
	"github.com/theo-henon/lcloud/internal/indexer"
	"github.com/theo-henon/lcloud/internal/monitoring"
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

	if err := db.AutoMigrate(&auth.User{}, &auth.RefreshToken{}, &volume.Volume{}); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	authService := auth.NewService(db, cfg.JWTSecret, cfg.JWTExpiryHours, cfg.RefreshTokenExpiryDays)
	if err := authService.SeedAdmin(cfg.AdminEmail, cfg.AdminPassword); err != nil {
		log.Fatalf("seed admin: %v", err)
	}

	diskRegistry := volume.NewDiskRegistry(cfg)
	indexManager := indexer.NewIndexManager()
	volumeService := volume.NewService(db, diskRegistry, indexManager)
	monitoringService := monitoring.NewService(volumeService, diskRegistry)
	fileService := volume.NewFileService(volumeService, indexManager, cfg.MaxUploadBytes, monitoringService.StatsCache())

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
		IndexManager:      indexManager,
		MaxUploadBytes:    cfg.MaxUploadBytes,
		StaticFS:          staticFS,
		GinMode:           cfg.GinMode,
	})

	addr := ":" + cfg.AppPort
	log.Printf("lcloud listening on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server: %v", err)
	}
}
