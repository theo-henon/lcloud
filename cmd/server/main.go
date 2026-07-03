package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"time"

	"github.com/theo-henon/lcloud/internal/api"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/config"
	"github.com/theo-henon/lcloud/internal/dashboard"
	"github.com/theo-henon/lcloud/internal/indexer"
	"github.com/theo-henon/lcloud/internal/monitoring"
	"github.com/theo-henon/lcloud/internal/plugin"
	"github.com/theo-henon/lcloud/internal/protocols"
	protocolftp "github.com/theo-henon/lcloud/internal/protocols/ftp"
	"github.com/theo-henon/lcloud/internal/settings"
	"github.com/theo-henon/lcloud/internal/task"
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

	if err := volume.PrepareProtocolsColumn(db); err != nil {
		log.Fatalf("migrate protocols column: %v", err)
	}

	if err := db.AutoMigrate(
		&auth.User{},
		&auth.RefreshToken{},
		&volume.Volume{},
		&volume.VolumeDeletionRequest{},
		&settings.InstanceSettings{},
		&plugin.Plugin{},
		&plugin.PluginLogEntry{},
		&task.Task{},
		&task.TaskRun{},
		&dashboard.UserDashboardLayout{},
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
	monitoringService := monitoring.NewService(volumeService, diskRegistry, authService, settingsService)
	fileService := volume.NewFileService(volumeService, indexManager, cfg.MaxUploadBytes, monitoringService.StatsCache())
	trashService := volume.NewTrashService(volumeService, indexManager, monitoringService.StatsCache())

	pluginService := plugin.NewService(db, cfg.PluginsPath, volumeService, nil)
	volumeService.SetEventPublisher(pluginService.Publisher())

	macroOps := volume.NewMacroOps(volumeService, indexManager, monitoringService.StatsCache(), pluginService.Publisher())
	taskExecutor := task.NewExecutor(macroOps, monitoringService, volumeService, pluginService)
	taskService := task.NewService(db, volumeService, authService, taskExecutor, pluginService)
	taskScheduler, err := task.NewScheduler(taskService)
	if err != nil {
		log.Fatalf("task scheduler: %v", err)
	}
	taskService.SetScheduler(taskScheduler)
	fileService.SetEventPublisher(pluginService.Publisher())
	trashService.SetEventPublisher(pluginService.Publisher())

	dashboardService := dashboard.NewService(db)

	protocolGateway := protocols.NewGateway(authService, volumeService, fileService, settingsService, cfg.FTPPort)
	ftpServer := protocolftp.NewServer(
		protocolGateway,
		fmt.Sprintf(":%d", cfg.FTPPort),
		cfg.FTPPasvAddress,
		cfg.FTPPasvMin,
		cfg.FTPPasvMax,
	)
	if err := ftpServer.Start(); err != nil {
		log.Fatalf("ftp server: %v", err)
	}
	defer ftpServer.Stop()

	ctx := context.Background()
	if err := pluginService.Start(ctx); err != nil {
		log.Fatalf("plugin startup: %v", err)
	}
	defer pluginService.Stop()

	enabledTasks, err := taskService.ListEnabled()
	if err != nil {
		log.Fatalf("load tasks: %v", err)
	}
	if err := taskScheduler.LoadAll(ctx, enabledTasks); err != nil {
		log.Fatalf("schedule tasks: %v", err)
	}
	taskScheduler.Start()
	defer func() {
		_ = taskScheduler.Shutdown()
	}()

	staticFS, err := fs.Sub(staticEmbed, "static")
	if err != nil {
		log.Fatalf("static fs: %v", err)
	}

	router := api.NewRouter(api.RouterConfig{
		AuthService:       authService,
		DiskRegistry:      diskRegistry,
		VolumeService:     volumeService,
		FileService:       fileService,
		TrashService:      trashService,
		MonitoringService: monitoringService,
		SettingsService:   settingsService,
		DashboardService:  dashboardService,
		PluginService:     pluginService,
		TaskService:       taskService,
		IndexManager:      indexManager,
		MaxUploadBytes:    cfg.MaxUploadBytes,
		FTPPort:           cfg.FTPPort,
		ProtocolGateway:   protocolGateway,
		StaticFS:          staticFS,
		GinMode:           cfg.GinMode,
	})

	addr := ":" + cfg.AppPort
	server := &http.Server{Addr: addr, Handler: router}

	go runTrashSweeper(context.Background(), trashService, cfg.TrashRetentionDays)

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		ftpServer.Stop()
		_ = taskScheduler.Shutdown()
		pluginService.Stop()
		_ = server.Close()
	}()

	log.Printf("lcloud listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}

func runTrashSweeper(ctx context.Context, trash *volume.TrashService, retentionDays int) {
	if retentionDays <= 0 {
		return
	}
	sweep := func() {
		count, err := trash.PurgeOlderThanAllVolumes(ctx, retentionDays)
		if err != nil {
			log.Printf("trash sweeper: %v", err)
			return
		}
		if count > 0 {
			log.Printf("trash sweeper: purged %d items across volumes (retention %d days)", count, retentionDays)
		}
	}
	go func() {
		sweep()
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				sweep()
			}
		}
	}()
}
