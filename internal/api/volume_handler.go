package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/settings"
	"github.com/theo-henon/lcloud/internal/volume"
	"github.com/theo-henon/lcloud/pkg/httputil"
)

type VolumeHandler struct {
	volumes  *volume.Service
	settings *settings.Service
}

func NewVolumeHandler(volumes *volume.Service, settingsService *settings.Service) *VolumeHandler {
	return &VolumeHandler{volumes: volumes, settings: settingsService}
}

type createVolumeRequest struct {
	Name       string         `json:"name" binding:"required"`
	DiskPath   string         `json:"disk_path" binding:"required"`
	QuotaBytes int64          `json:"quota_bytes"`
	Filters    volume.Filters `json:"filters"`
}

type patchVolumeRequest struct {
	Name       *string         `json:"name"`
	QuotaBytes *int64          `json:"quota_bytes"`
	Filters    *volume.Filters `json:"filters"`
}

func (h *VolumeHandler) List(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httputil.Unauthorized(c, "unauthorized")
		return
	}

	volumes, err := h.volumes.List(claims)
	if err != nil {
		httputil.InternalError(c, "unable to list volumes")
		return
	}
	for i := range volumes {
		masked, err := h.maskVolumeIfNeeded(claims, volumes[i])
		if err != nil {
			httputil.InternalError(c, "unable to load settings")
			return
		}
		volumes[i] = masked
	}
	httputil.JSON(c, http.StatusOK, gin.H{"volumes": volumes})
}

func (h *VolumeHandler) Create(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httputil.Unauthorized(c, "unauthorized")
		return
	}

	var req createVolumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "invalid request body")
		return
	}

	vol, err := h.volumes.Create(claims, volume.CreateVolumeInput{
		Name:       req.Name,
		DiskPath:   req.DiskPath,
		QuotaBytes: req.QuotaBytes,
		Filters:    req.Filters,
	})
	if err != nil {
		mapVolumeError(c, err)
		return
	}
	masked, err := h.maskVolumeIfNeeded(claims, *vol)
	if err != nil {
		httputil.InternalError(c, "unable to load settings")
		return
	}
	httputil.JSON(c, http.StatusCreated, masked)
}

func (h *VolumeHandler) Get(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httputil.Unauthorized(c, "unauthorized")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.BadRequest(c, "invalid volume id")
		return
	}

	vol, err := h.volumes.Get(claims, id)
	if err != nil {
		mapVolumeError(c, err)
		return
	}
	masked, err := h.maskVolumeIfNeeded(claims, *vol)
	if err != nil {
		httputil.InternalError(c, "unable to load settings")
		return
	}
	httputil.JSON(c, http.StatusOK, masked)
}

func (h *VolumeHandler) Patch(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httputil.Unauthorized(c, "unauthorized")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.BadRequest(c, "invalid volume id")
		return
	}

	var req patchVolumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "invalid request body")
		return
	}

	vol, err := h.volumes.Patch(claims, id, volume.PatchVolumeInput{
		Name:       req.Name,
		QuotaBytes: req.QuotaBytes,
		Filters:    req.Filters,
	})
	if err != nil {
		mapVolumeError(c, err)
		return
	}
	masked, err := h.maskVolumeIfNeeded(claims, *vol)
	if err != nil {
		httputil.InternalError(c, "unable to load settings")
		return
	}
	httputil.JSON(c, http.StatusOK, masked)
}

func (h *VolumeHandler) Delete(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httputil.Unauthorized(c, "unauthorized")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.BadRequest(c, "invalid volume id")
		return
	}

	force := c.Query("force") == "true"
	if err := h.volumes.Delete(claims, id, force); err != nil {
		mapVolumeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func mapVolumeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, volume.ErrVolumeNotFound):
		httputil.Error(c, http.StatusNotFound, "NOT_FOUND", "volume not found")
	case errors.Is(err, volume.ErrForbidden):
		httputil.Forbidden(c, "forbidden")
	case errors.Is(err, volume.ErrDiskNotFound):
		httputil.Unprocessable(c, "DISK_NOT_FOUND", "disk not found")
	case errors.Is(err, volume.ErrVolumeNameTaken):
		httputil.Conflict(c, "volume name already taken")
	case errors.Is(err, volume.ErrInvalidFilter):
		httputil.Unprocessable(c, "INVALID_FILTER", "invalid filter configuration")
	case errors.Is(err, volume.ErrVolumeNotEmpty):
		httputil.Unprocessable(c, "VOLUME_NOT_EMPTY", "volume is not empty")
	default:
		httputil.InternalError(c, "volume operation failed")
	}
}

func (h *VolumeHandler) maskVolumeIfNeeded(claims *auth.Claims, vol volume.Volume) (volume.Volume, error) {
	mask, err := h.settings.ShouldMaskFor(claims)
	if err != nil {
		return volume.Volume{}, err
	}
	if mask {
		return maskVolume(vol), nil
	}
	return vol, nil
}
