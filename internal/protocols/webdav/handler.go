package webdav

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/protocols"
	"github.com/theo-henon/lcloud/internal/volume"
	gowebdav "golang.org/x/net/webdav"
)

func RegisterRoutes(router *gin.Engine, gateway *protocols.Gateway) {
	router.Any("/dav/volumes/:id/*path", serveWebDAV(gateway))
	router.Any("/dav/volumes/:id", serveWebDAV(gateway))
}

func serveWebDAV(gateway *protocols.Gateway) gin.HandlerFunc {
	return func(c *gin.Context) {
		volumeID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		email, password, ok := c.Request.BasicAuth()
		if !ok {
			c.Header("WWW-Authenticate", `Basic realm="lcloud WebDAV"`)
			c.Status(http.StatusUnauthorized)
			return
		}

		claims, parsedID, err := gateway.WebDAVLogin(c.Request, email, password)
		if err != nil {
			if errors.Is(err, auth.ErrInvalidCredentials) || errors.Is(err, auth.ErrUserDisabled) {
				c.Header("WWW-Authenticate", `Basic realm="lcloud WebDAV"`)
				c.Status(http.StatusUnauthorized)
				return
			}
			if errors.Is(err, volume.ErrVolumeNotFound) {
				c.Status(http.StatusNotFound)
				return
			}
			if errors.Is(err, volume.ErrProtocolsDisabled) || errors.Is(err, volume.ErrForbidden) {
				c.Status(http.StatusForbidden)
				return
			}
			c.Status(http.StatusInternalServerError)
			return
		}
		if parsedID != volumeID {
			c.Status(http.StatusNotFound)
			return
		}

		prefix := "/dav/volumes/" + volumeID.String() + "/"
		handler := &gowebdav.Handler{
			Prefix:     prefix,
			FileSystem: newVolumeFS(gateway.Files(), claims, volumeID),
		}

		pathSuffix := c.Param("path")
		if pathSuffix == "" {
			if !strings.HasSuffix(c.Request.URL.Path, "/") {
				c.Redirect(http.StatusMovedPermanently, prefix)
				return
			}
		}

		handler.ServeHTTP(c.Writer, c.Request)
	}
}
