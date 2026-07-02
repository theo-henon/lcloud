package api

import (
	"errors"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/indexer"
	"github.com/theo-henon/lcloud/internal/monitoring"
	"github.com/theo-henon/lcloud/internal/plugin"
	"github.com/theo-henon/lcloud/internal/settings"
	"github.com/theo-henon/lcloud/internal/task"
	"github.com/theo-henon/lcloud/internal/volume"
	"github.com/theo-henon/lcloud/pkg/httputil"
)

type AuthHandler struct {
	auth *auth.Service
}

func NewAuthHandler(authService *auth.Service) *AuthHandler {
	return &AuthHandler{auth: authService}
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "invalid request body")
		return
	}

	result, err := h.auth.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			httputil.Unauthorized(c, "invalid credentials")
			return
		}
		httputil.InternalError(c, "unable to login")
		return
	}

	httputil.JSON(c, http.StatusOK, result)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "invalid request body")
		return
	}

	result, err := h.auth.Refresh(req.RefreshToken)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidToken) || errors.Is(err, auth.ErrExpiredToken) || errors.Is(err, auth.ErrRevokedToken) {
			httputil.Unauthorized(c, "invalid refresh token")
			return
		}
		httputil.InternalError(c, "unable to refresh token")
		return
	}

	httputil.JSON(c, http.StatusOK, result)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "invalid request body")
		return
	}

	if err := h.auth.Logout(req.RefreshToken); err != nil {
		if errors.Is(err, auth.ErrInvalidToken) || errors.Is(err, auth.ErrExpiredToken) || errors.Is(err, auth.ErrRevokedToken) {
			httputil.Unauthorized(c, "invalid refresh token")
			return
		}
		httputil.InternalError(c, "unable to logout")
		return
	}

	httputil.JSON(c, http.StatusOK, gin.H{"status": "ok"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httputil.Unauthorized(c, "unauthorized")
		return
	}

	user, err := h.auth.GetUserByID(claims.UserID)
	if err != nil {
		httputil.Unauthorized(c, "unauthorized")
		return
	}

	httputil.JSON(c, http.StatusOK, auth.ToUserResponse(user))
}

type AdminHandler struct {
	auth *auth.Service
}

func NewAdminHandler(authService *auth.Service) *AdminHandler {
	return &AdminHandler{auth: authService}
}

type createUserRequest struct {
	Email    string    `json:"email" binding:"required,email"`
	Password string    `json:"password" binding:"required,min=8"`
	Role     auth.Role `json:"role"`
}

func (h *AdminHandler) CreateUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "invalid request body")
		return
	}

	user, err := h.auth.CreateUser(req.Email, req.Password, req.Role)
	if err != nil {
		if errors.Is(err, auth.ErrEmailTaken) {
			httputil.Conflict(c, "email already taken")
			return
		}
		if errors.Is(err, auth.ErrInvalidRole) {
			httputil.BadRequest(c, "role must be admin or user")
			return
		}
		httputil.InternalError(c, "unable to create user")
		return
	}

	httputil.JSON(c, http.StatusCreated, auth.ToUserResponse(user))
}

type RouterConfig struct {
	AuthService       *auth.Service
	DiskRegistry      *volume.DiskRegistry
	VolumeService     *volume.Service
	FileService       *volume.FileService
	MonitoringService *monitoring.Service
	SettingsService   *settings.Service
	PluginService     *plugin.Service
	TaskService       *task.Service
	IndexManager      *indexer.IndexManager
	MaxUploadBytes    int64
	StaticFS          fs.FS
	GinMode           string
}

func NewRouter(cfg RouterConfig) *gin.Engine {
	mode := cfg.GinMode
	if mode == "" {
		mode = gin.ReleaseMode
	}
	gin.SetMode(mode)

	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())
	// Buffer multipart parsing in memory; actual upload size is enforced in FileService.
	router.MaxMultipartMemory = 32 << 20

	authHandler := NewAuthHandler(cfg.AuthService)
	adminHandler := NewAdminHandler(cfg.AuthService)
	diskHandler := NewDiskHandler(cfg.DiskRegistry, cfg.SettingsService)
	volumeHandler := NewVolumeHandler(cfg.VolumeService, cfg.SettingsService)
	fileHandler := NewFileHandler(cfg.FileService)
	monitoringHandler := NewMonitoringHandler(cfg.MonitoringService)
	searchHandler := NewSearchHandler(cfg.VolumeService, cfg.IndexManager)
	settingsHandler := NewSettingsHandler(cfg.SettingsService)
	adminSettingsHandler := NewAdminSettingsHandler(cfg.SettingsService)
	pluginHandler := NewPluginHandler(cfg.PluginService)
	taskHandler := NewTaskHandler(cfg.TaskService)

	api := router.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			httputil.JSON(c, http.StatusOK, gin.H{"status": "ok"})
		})

		authGroup := api.Group("/auth")
		{
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/refresh", authHandler.Refresh)
			authGroup.POST("/logout", authHandler.Logout)
			authGroup.GET("/me", auth.AuthMiddleware(cfg.AuthService), authHandler.Me)
		}

		admin := api.Group("/admin", auth.AuthMiddleware(cfg.AuthService), auth.RequireAdmin())
		{
			admin.POST("/users", adminHandler.CreateUser)
			admin.PATCH("/settings", adminSettingsHandler.Patch)
		}

		protected := api.Group("", auth.AuthMiddleware(cfg.AuthService))
		{
			protected.GET("/settings", settingsHandler.Get)
			protected.GET("/disks", diskHandler.List)
			protected.GET("/search", searchHandler.SearchAll)

			protected.GET("/volumes", volumeHandler.List)
			protected.POST("/volumes", auth.RequireAdmin(), volumeHandler.Create)
			protected.GET("/volumes/:id", volumeHandler.Get)
			protected.PATCH("/volumes/:id", volumeHandler.Patch)
			protected.DELETE("/volumes/:id", volumeHandler.Delete)

			protected.GET("/volumes/:id/files", fileHandler.List)
			protected.POST("/volumes/:id/files/directories", fileHandler.CreateDirectory)
			protected.POST("/volumes/:id/files", fileHandler.Upload)
			protected.GET("/volumes/:id/files/content", fileHandler.Download)
			protected.GET("/volumes/:id/files/thumbnail", fileHandler.Thumbnail)
			protected.DELETE("/volumes/:id/files", fileHandler.Delete)
			protected.GET("/volumes/:id/search", searchHandler.Search)

			protected.GET("/monitoring/overview", monitoringHandler.Overview)
			protected.GET("/monitoring/volumes/:id/stats", monitoringHandler.VolumeStats)
			protected.POST("/monitoring/volumes/:id/stats/refresh", monitoringHandler.RefreshVolumeStats)

			protected.GET("/plugins", pluginHandler.List)
			protected.GET("/plugins/logs", pluginHandler.Logs)
			protected.GET("/plugins/:id", pluginHandler.Get)
			protected.PATCH("/plugins/:id", auth.RequireAdmin(), pluginHandler.Patch)

			protected.GET("/tasks", taskHandler.List)
			protected.POST("/tasks", taskHandler.Create)
			protected.GET("/tasks/:id", taskHandler.Get)
			protected.PATCH("/tasks/:id", taskHandler.Patch)
			protected.DELETE("/tasks/:id", taskHandler.Delete)
			protected.POST("/tasks/:id/run", taskHandler.Run)
			protected.GET("/tasks/:id/runs", taskHandler.Runs)
		}
	}

	if cfg.StaticFS != nil {
		router.NoRoute(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api") {
				c.Status(http.StatusNotFound)
				return
			}

			path := strings.TrimPrefix(c.Request.URL.Path, "/")
			if path == "" {
				path = "index.html"
			}

			if _, err := fs.Stat(cfg.StaticFS, path); err != nil {
				path = "index.html"
			}

			data, err := fs.ReadFile(cfg.StaticFS, path)
			if err != nil {
				c.Status(http.StatusNotFound)
				return
			}

			contentType := mime.TypeByExtension(filepath.Ext(path))
			if contentType == "" {
				contentType = "application/octet-stream"
			}

			c.Data(http.StatusOK, contentType, data)
		})
	}

	return router
}
