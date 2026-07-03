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
	handler := serveWebDAV(gateway)
	routes := []string{
		"/dav/volumes/:id/*path",
		"/dav/volumes/:id",
	}
	// Gin router.Any only covers standard HTTP methods; WebDAV needs PROPFIND, MKCOL, MOVE, etc.
	methods := []string{
		http.MethodGet,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPost,
		http.MethodHead,
		http.MethodOptions,
		"PROPFIND",
		"MKCOL",
		"MOVE",
		"COPY",
		"LOCK",
		"UNLOCK",
	}
	for _, route := range routes {
		for _, method := range methods {
			router.Handle(method, route, handler)
		}
	}
}

func serveWebDAV(gateway *protocols.Gateway) gin.HandlerFunc {
	return func(c *gin.Context) {
		volumeID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		// macOS Finder probes OPTIONS before auth; answer with DAV capabilities.
		if c.Request.Method == http.MethodOptions {
			if _, _, ok := c.Request.BasicAuth(); !ok {
				writeDAVOptionsHeaders(c)
				c.Status(http.StatusOK)
				return
			}
		}

		email, password, ok := c.Request.BasicAuth()
		if !ok {
			writeDAVOptionsHeaders(c)
			c.Header("WWW-Authenticate", `Basic realm="lcloud WebDAV"`)
			c.Status(http.StatusUnauthorized)
			return
		}

		claims, parsedID, err := gateway.WebDAVLogin(c.Request, email, password)
		if err != nil {
			if errors.Is(err, auth.ErrInvalidCredentials) || errors.Is(err, auth.ErrUserDisabled) {
				writeDAVOptionsHeaders(c)
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
			LockSystem: gowebdav.NewMemLS(),
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

func writeDAVOptionsHeaders(c *gin.Context) {
	c.Header("DAV", "1, 2")
	c.Header("Allow", "OPTIONS, GET, HEAD, PUT, DELETE, PROPFIND, MKCOL, MOVE")
	c.Header("MS-Author-Via", "DAV")
}
