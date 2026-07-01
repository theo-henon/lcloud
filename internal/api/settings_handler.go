package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/theo-henon/lcloud/internal/settings"
	"github.com/theo-henon/lcloud/pkg/httputil"
)

type SettingsHandler struct {
	settings *settings.Service
}

func NewSettingsHandler(settingsService *settings.Service) *SettingsHandler {
	return &SettingsHandler{settings: settingsService}
}

func (h *SettingsHandler) Get(c *gin.Context) {
	public, err := h.settings.Public()
	if err != nil {
		httputil.InternalError(c, "unable to load settings")
		return
	}
	httputil.JSON(c, http.StatusOK, public)
}

type AdminSettingsHandler struct {
	settings *settings.Service
}

func NewAdminSettingsHandler(settingsService *settings.Service) *AdminSettingsHandler {
	return &AdminSettingsHandler{settings: settingsService}
}

func (h *AdminSettingsHandler) Patch(c *gin.Context) {
	var input settings.UpdateSettingsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.BadRequest(c, "invalid request body")
		return
	}
	if input.MaskDiskNames == nil {
		httputil.BadRequest(c, "mask_disk_names is required")
		return
	}

	public, err := h.settings.Update(input)
	if err != nil {
		httputil.InternalError(c, "unable to update settings")
		return
	}
	httputil.JSON(c, http.StatusOK, public)
}
