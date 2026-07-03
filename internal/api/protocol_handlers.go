package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/volume"
	"github.com/theo-henon/lcloud/pkg/httputil"
)

type ProtocolHandler struct {
	volumes *volume.Service
	ftpPort int
}

func NewProtocolHandler(volumes *volume.Service, ftpPort int) *ProtocolHandler {
	return &ProtocolHandler{volumes: volumes, ftpPort: ftpPort}
}

type patchProtocolsRequest struct {
	WebDAV *volume.ProtocolToggle `json:"webdav"`
	FTP    *volume.ProtocolToggle `json:"ftp"`
}

func (h *ProtocolHandler) Get(c *gin.Context) {
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

	host := requestHost(c)
	info, err := h.volumes.GetProtocols(claims, volumeID, host, h.ftpPort)
	if err != nil {
		switch {
		case errors.Is(err, volume.ErrVolumeNotFound):
			httputil.Error(c, http.StatusNotFound, "NOT_FOUND", "volume not found")
		case errors.Is(err, volume.ErrForbidden):
			httputil.Forbidden(c, "forbidden")
		default:
			httputil.InternalError(c, "unable to load protocol settings")
		}
		return
	}

	httputil.JSON(c, http.StatusOK, info)
}

func (h *ProtocolHandler) Patch(c *gin.Context) {
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

	var req patchProtocolsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "invalid request body")
		return
	}
	if req.WebDAV == nil && req.FTP == nil {
		httputil.BadRequest(c, "at least one protocol field required")
		return
	}

	info, err := h.volumes.UpdateProtocols(claims, volumeID, volume.UpdateProtocolsInput{
		WebDAV: req.WebDAV,
		FTP:    req.FTP,
	})
	if err != nil {
		switch {
		case errors.Is(err, volume.ErrVolumeNotFound):
			httputil.Error(c, http.StatusNotFound, "NOT_FOUND", "volume not found")
		case errors.Is(err, volume.ErrForbidden):
			httputil.Forbidden(c, "forbidden")
		case errors.Is(err, volume.ErrInvalidFilter):
			httputil.BadRequest(c, "invalid request body")
		default:
			httputil.InternalError(c, "unable to update protocol settings")
		}
		return
	}

	host := requestHost(c)
	info.Connection.WebDAVURL = buildWebDAVURL(host, volumeID)
	info.Connection.FTPHost = host
	info.Connection.FTPPort = h.ftpPort

	httputil.JSON(c, http.StatusOK, info)
}

func requestHost(c *gin.Context) string {
	if forwarded := c.GetHeader("X-Forwarded-Host"); forwarded != "" {
		if host := strings.Split(forwarded, ",")[0]; host != "" {
			return strings.TrimSpace(strings.Split(host, ":")[0])
		}
	}
	if host := c.Request.Host; host != "" {
		return strings.Split(host, ":")[0]
	}
	return "localhost"
}

func buildWebDAVURL(host string, volumeID uuid.UUID) string {
	return "http://" + host + "/dav/volumes/" + volumeID.String() + "/"
}
