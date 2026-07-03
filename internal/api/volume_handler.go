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
	DiskPath   string         `json:"disk_path"`
	DiskID     string         `json:"disk_id"`
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

	var ownerID *uuid.UUID
	if raw := c.Query("owner_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			httputil.BadRequest(c, "invalid owner_id")
			return
		}
		if claims.Role != auth.RoleAdmin && id != claims.UserID {
			httputil.Forbidden(c, "forbidden")
			return
		}
		ownerID = &id
	}

	volumes, err := h.volumes.List(claims, ownerID)
	if err != nil {
		httputil.InternalError(c, "unable to list volumes")
		return
	}

	pendingIDs, err := h.volumes.PendingDeletionVolumeIDs(claims.UserID)
	if err != nil {
		httputil.InternalError(c, "unable to list volumes")
		return
	}

	var emailByOwner map[uuid.UUID]string
	if claims.Role == auth.RoleAdmin {
		ownerIDs := make([]uuid.UUID, len(volumes))
		for i := range volumes {
			ownerIDs[i] = volumes[i].OwnerID
		}
		var err error
		emailByOwner, err = h.volumes.OwnerEmailsByIDs(ownerIDs)
		if err != nil {
			httputil.InternalError(c, "unable to list volumes")
			return
		}
	}

	items := make([]gin.H, len(volumes))
	for i := range volumes {
		masked, err := h.maskVolumeIfNeeded(claims, volumes[i])
		if err != nil {
			httputil.InternalError(c, "unable to load settings")
			return
		}
		item := gin.H{
			"id":                        masked.ID,
			"name":                      masked.Name,
			"owner_id":                  masked.OwnerID,
			"disk_path":                 masked.DiskPath,
			"root_path":                 masked.RootPath,
			"quota_bytes":               masked.QuotaBytes,
			"used_bytes":                masked.UsedBytes,
			"filters":                   masked.Filters,
			"created_at":                masked.CreatedAt,
			"updated_at":                masked.UpdatedAt,
			"deletion_request_pending":  pendingIDs[masked.ID],
		}
		if emailByOwner != nil {
			if email := emailByOwner[masked.OwnerID]; email != "" {
				item["owner_email"] = email
			}
		}
		items[i] = item
	}
	httputil.JSON(c, http.StatusOK, gin.H{"volumes": items})
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
	if req.DiskPath == "" && req.DiskID == "" {
		httputil.BadRequest(c, "disk_path or disk_id is required")
		return
	}

	vol, err := h.volumes.Create(claims, volume.CreateVolumeInput{
		Name:       req.Name,
		DiskPath:   req.DiskPath,
		DiskID:     req.DiskID,
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

func (h *VolumeHandler) RequestDeletion(c *gin.Context) {
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

	if err := h.volumes.RequestDeletion(claims, id); err != nil {
		mapVolumeError(c, err)
		return
	}
	httputil.JSON(c, http.StatusCreated, gin.H{"status": "pending"})
}

func (h *VolumeHandler) ListDeletionRequests(c *gin.Context) {
	requests, err := h.volumes.ListPendingDeletionRequests()
	if err != nil {
		httputil.InternalError(c, "unable to list deletion requests")
		return
	}
	if requests == nil {
		requests = []volume.DeletionRequestResponse{}
	}
	httputil.JSON(c, http.StatusOK, gin.H{"requests": requests})
}

func (h *VolumeHandler) DismissDeletionRequest(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.BadRequest(c, "invalid request id")
		return
	}

	if err := h.volumes.DismissDeletionRequest(id); err != nil {
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
	case errors.Is(err, volume.ErrDeletionRequestExists):
		httputil.Conflict(c, "deletion request already pending")
	case errors.Is(err, volume.ErrDeletionRequestNotFound):
		httputil.Error(c, http.StatusNotFound, "NOT_FOUND", "deletion request not found")
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
