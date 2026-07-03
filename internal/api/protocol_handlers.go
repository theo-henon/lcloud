package api

import (
	"errors"
	"net"
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

	h.enrichConnection(c, info, volumeID)
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

	h.enrichConnection(c, info, volumeID)
	httputil.JSON(c, http.StatusOK, info)
}

func (h *ProtocolHandler) enrichConnection(c *gin.Context, info *volume.ProtocolsInfo, volumeID uuid.UUID) {
	authority := requestAuthority(c)
	info.Connection.WebDAVURL = buildWebDAVURL(c, authority, volumeID)
	info.Connection.FTPHost = requestHost(c)
	info.Connection.FTPPort = h.ftpPort
}

func requestAuthority(c *gin.Context) string {
	if forwarded := c.GetHeader("X-Forwarded-Host"); forwarded != "" {
		if host := strings.TrimSpace(strings.Split(forwarded, ",")[0]); host != "" {
			return host
		}
	}
	if c.Request.Host != "" {
		return c.Request.Host
	}
	return "localhost:8080"
}

func requestHost(c *gin.Context) string {
	authority := requestAuthority(c)
	host, _, err := net.SplitHostPort(authority)
	if err != nil {
		return authority
	}
	return host
}

func buildWebDAVURL(c *gin.Context, authority string, volumeID uuid.UUID) string {
	scheme := "http"
	if proto := c.GetHeader("X-Forwarded-Proto"); proto == "https" {
		scheme = "https"
	}
	return scheme + "://" + authority + "/dav/volumes/" + volumeID.String() + "/"
}
