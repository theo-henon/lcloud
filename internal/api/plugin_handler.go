package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/plugin"
	"github.com/theo-henon/lcloud/pkg/httputil"
)

type PluginHandler struct {
	plugins *plugin.Service
}

func NewPluginHandler(plugins *plugin.Service) *PluginHandler {
	return &PluginHandler{plugins: plugins}
}

type patchPluginRequest struct {
	Enabled bool `json:"enabled"`
}

func (h *PluginHandler) List(c *gin.Context) {
	records, err := h.plugins.List()
	if err != nil {
		httputil.InternalError(c, "unable to list plugins")
		return
	}
	httputil.JSON(c, http.StatusOK, gin.H{"plugins": records})
}

func (h *PluginHandler) Get(c *gin.Context) {
	record, err := h.plugins.Get(c.Param("id"))
	if err != nil {
		if errors.Is(err, plugin.ErrPluginNotFound) {
			httputil.JSON(c, http.StatusNotFound, gin.H{"error": "PLUGIN_NOT_FOUND"})
			return
		}
		httputil.InternalError(c, "unable to get plugin")
		return
	}
	httputil.JSON(c, http.StatusOK, record)
}

func (h *PluginHandler) Patch(c *gin.Context) {
	var req patchPluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "invalid request body")
		return
	}

	if err := h.plugins.SetEnabled(c.Request.Context(), c.Param("id"), req.Enabled); err != nil {
		switch {
		case errors.Is(err, plugin.ErrPluginNotFound):
			httputil.JSON(c, http.StatusNotFound, gin.H{"error": "PLUGIN_NOT_FOUND"})
		case errors.Is(err, plugin.ErrPluginStartFailed):
			httputil.JSON(c, http.StatusBadRequest, gin.H{"error": "PLUGIN_START_FAILED", "message": err.Error()})
		default:
			httputil.InternalError(c, "unable to update plugin")
		}
		return
	}

	record, err := h.plugins.Get(c.Param("id"))
	if err != nil {
		httputil.InternalError(c, "unable to get plugin")
		return
	}
	httputil.JSON(c, http.StatusOK, record)
}

func (h *PluginHandler) Logs(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httputil.Unauthorized(c, "unauthorized")
		return
	}

	filter := plugin.LogFilter{PluginID: c.Query("plugin_id")}
	if volumeID := c.Query("volume_id"); volumeID != "" {
		id, err := uuid.Parse(volumeID)
		if err != nil {
			httputil.BadRequest(c, "invalid volume_id")
			return
		}
		filter.VolumeID = &id
	}
	if limitRaw := c.Query("limit"); limitRaw != "" {
		limit, err := strconv.Atoi(limitRaw)
		if err != nil {
			httputil.BadRequest(c, "invalid limit")
			return
		}
		filter.Limit = limit
	}

	entries, err := h.plugins.ListLogs(c.Request.Context(), claims, filter)
	if err != nil {
		httputil.InternalError(c, "unable to list plugin logs")
		return
	}
	httputil.JSON(c, http.StatusOK, gin.H{"entries": entries})
}
