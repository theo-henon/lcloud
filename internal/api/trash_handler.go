package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/volume"
	"github.com/theo-henon/lcloud/pkg/httputil"
)

type TrashHandler struct {
	trash *volume.TrashService
}

func NewTrashHandler(trash *volume.TrashService) *TrashHandler {
	return &TrashHandler{trash: trash}
}

func (h *TrashHandler) List(c *gin.Context) {
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

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	listing, err := h.trash.List(claims, volumeID, limit, offset)
	if err != nil {
		mapTrashError(c, err)
		return
	}
	httputil.JSON(c, http.StatusOK, listing)
}

func (h *TrashHandler) Restore(c *gin.Context) {
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
	fileID := c.Param("fileId")
	if fileID == "" {
		httputil.BadRequest(c, "file id is required")
		return
	}

	entry, err := h.trash.Restore(claims, volumeID, fileID)
	if err != nil {
		mapTrashError(c, err)
		return
	}
	httputil.JSON(c, http.StatusOK, entry)
}

func (h *TrashHandler) PurgeOne(c *gin.Context) {
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
	fileID := c.Param("fileId")
	if fileID == "" {
		httputil.BadRequest(c, "file id is required")
		return
	}

	if err := h.trash.PurgeOne(claims, volumeID, fileID); err != nil {
		mapTrashError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *TrashHandler) Empty(c *gin.Context) {
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

	purged, err := h.trash.Empty(claims, volumeID)
	if err != nil {
		mapTrashError(c, err)
		return
	}
	httputil.JSON(c, http.StatusOK, gin.H{"purged": purged})
}

func mapTrashError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, volume.ErrVolumeNotFound), errors.Is(err, volume.ErrTrashItemNotFound), errors.Is(err, volume.ErrFileNotFound):
		httputil.Error(c, http.StatusNotFound, "NOT_FOUND", "resource not found")
	case errors.Is(err, volume.ErrForbidden):
		httputil.Forbidden(c, "forbidden")
	case errors.Is(err, volume.ErrPathOccupied):
		httputil.Error(c, http.StatusConflict, "PATH_OCCUPIED", "path already exists")
	case errors.Is(err, volume.ErrPathTraversal):
		httputil.Unprocessable(c, "PATH_TRAVERSAL", "invalid path")
	default:
		httputil.InternalError(c, "trash operation failed")
	}
}
