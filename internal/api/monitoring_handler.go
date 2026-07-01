package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/monitoring"
	"github.com/theo-henon/lcloud/internal/volume"
	"github.com/theo-henon/lcloud/pkg/httputil"
)

type MonitoringHandler struct {
	monitoring *monitoring.Service
}

func NewMonitoringHandler(monitoringService *monitoring.Service) *MonitoringHandler {
	return &MonitoringHandler{monitoring: monitoringService}
}

func (h *MonitoringHandler) Overview(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httputil.Unauthorized(c, "unauthorized")
		return
	}

	overview, err := h.monitoring.GetOverview(c.Request.Context(), claims)
	if err != nil {
		httputil.InternalError(c, "unable to load monitoring overview")
		return
	}
	httputil.JSON(c, http.StatusOK, overview)
}

func (h *MonitoringHandler) VolumeStats(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httputil.Unauthorized(c, "unauthorized")
		return
	}

	volumeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.BadRequest(c, "invalid volume id")
		return
	}

	stats, err := h.monitoring.GetVolumeStats(c.Request.Context(), claims, volumeID)
	if err != nil {
		if errors.Is(err, volume.ErrVolumeNotFound) {
			httputil.Unprocessable(c, "VOLUME_NOT_FOUND", "volume not found")
			return
		}
		if errors.Is(err, volume.ErrForbidden) {
			httputil.Forbidden(c, "forbidden")
			return
		}
		httputil.InternalError(c, "unable to load volume stats")
		return
	}
	httputil.JSON(c, http.StatusOK, stats)
}

func (h *MonitoringHandler) RefreshVolumeStats(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httputil.Unauthorized(c, "unauthorized")
		return
	}

	volumeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.BadRequest(c, "invalid volume id")
		return
	}

	stats, err := h.monitoring.RefreshVolumeStats(c.Request.Context(), claims, volumeID)
	if err != nil {
		if errors.Is(err, volume.ErrVolumeNotFound) {
			httputil.Unprocessable(c, "VOLUME_NOT_FOUND", "volume not found")
			return
		}
		if errors.Is(err, volume.ErrForbidden) {
			httputil.Forbidden(c, "forbidden")
			return
		}
		httputil.InternalError(c, "unable to refresh volume stats")
		return
	}
	httputil.JSON(c, http.StatusOK, stats)
}
