package protocols

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/settings"
	"github.com/theo-henon/lcloud/internal/volume"
)

type Gateway struct {
	auth      *auth.Service
	volumes   *volume.Service
	files     *volume.FileService
	settings  *settings.Service
	rateLimit *authRateLimiter
	ftpPort   int
}

func NewGateway(
	authService *auth.Service,
	volumeService *volume.Service,
	fileService *volume.FileService,
	settingsService *settings.Service,
	ftpPort int,
) *Gateway {
	return &Gateway{
		auth:      authService,
		volumes:   volumeService,
		files:     fileService,
		settings:  settingsService,
		rateLimit: newAuthRateLimiter(),
		ftpPort:   ftpPort,
	}
}

func (g *Gateway) FTPPort() int {
	return g.ftpPort
}

func (g *Gateway) Files() *volume.FileService {
	return g.files
}

func (g *Gateway) Settings() *settings.Service {
	return g.settings
}

func (g *Gateway) WebDAVLogin(r *http.Request, email, password string) (*auth.Claims, uuid.UUID, error) {
	ip := clientIP(r.RemoteAddr)
	if !g.rateLimit.allow(ip) {
		return nil, uuid.Nil, auth.ErrInvalidCredentials
	}

	volumeID, err := parseWebDAVVolumeID(r.URL.Path)
	if err != nil {
		g.rateLimit.recordFailure(ip)
		return nil, uuid.Nil, err
	}

	claims, err := g.auth.ValidateCredentials(email, password)
	if err != nil {
		g.rateLimit.recordFailure(ip)
		return nil, uuid.Nil, err
	}

	if err := g.ensureAccess(claims, volumeID, volume.ProtocolWebDAV); err != nil {
		g.rateLimit.recordFailure(ip)
		return nil, uuid.Nil, err
	}

	return claims, volumeID, nil
}

func (g *Gateway) FTPLogin(remoteAddr, username, password string) (*auth.Claims, uuid.UUID, error) {
	ip := clientIP(remoteAddr)
	if !g.rateLimit.allow(ip) {
		return nil, uuid.Nil, auth.ErrInvalidCredentials
	}

	volumeID, err := uuid.Parse(strings.TrimSpace(username))
	if err != nil {
		g.rateLimit.recordFailure(ip)
		return nil, uuid.Nil, auth.ErrInvalidCredentials
	}

	claims, err := g.auth.AuthenticateForVolume(volumeID, password)
	if err != nil {
		g.rateLimit.recordFailure(ip)
		return nil, uuid.Nil, err
	}

	if err := g.ensureAccess(claims, volumeID, volume.ProtocolFTP); err != nil {
		g.rateLimit.recordFailure(ip)
		return nil, uuid.Nil, err
	}

	return claims, volumeID, nil
}

func (g *Gateway) ensureAccess(claims *auth.Claims, volumeID uuid.UUID, kind volume.ProtocolKind) error {
	vol, err := g.volumes.GetByID(volumeID)
	if err != nil {
		if errors.Is(err, volume.ErrVolumeNotFound) {
			return volume.ErrVolumeNotFound
		}
		return err
	}

	instanceWebDAV, err := g.settings.ProtocolsWebDAVEnabled()
	if err != nil {
		return err
	}
	instanceFTP, err := g.settings.ProtocolsFTPEnabled()
	if err != nil {
		return err
	}

	if !g.volumes.ProtocolEnabled(vol, kind, instanceWebDAV, instanceFTP) {
		return volume.ErrProtocolsDisabled
	}
	return g.volumes.AuthorizeVolume(claims, vol)
}

func parseWebDAVVolumeID(path string) (uuid.UUID, error) {
	path = strings.TrimPrefix(path, "/dav/volumes/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		return uuid.Nil, volume.ErrVolumeNotFound
	}
	return uuid.Parse(parts[0])
}

func requestHost(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-Host"); forwarded != "" {
		if host := strings.Split(forwarded, ",")[0]; host != "" {
			return strings.TrimSpace(host)
		}
	}
	if host := r.Host; host != "" {
		return strings.Split(host, ":")[0]
	}
	return "localhost"
}
