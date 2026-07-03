package volume

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/theo-henon/lcloud/internal/auth"
)

type ProtocolKind string

const (
	ProtocolWebDAV ProtocolKind = "webdav"
	ProtocolFTP    ProtocolKind = "ftp"
)

type ProtocolToggle struct {
	Enabled bool `json:"enabled"`
}

type ProtocolsConfig struct {
	WebDAV ProtocolToggle `json:"webdav"`
	FTP    ProtocolToggle `json:"ftp"`
}

func DefaultProtocolsConfig() ProtocolsConfig {
	return ProtocolsConfig{
		WebDAV: ProtocolToggle{Enabled: false},
		FTP:    ProtocolToggle{Enabled: false},
	}
}

type UpdateProtocolsInput struct {
	WebDAV *ProtocolToggle `json:"webdav"`
	FTP    *ProtocolToggle `json:"ftp"`
}

type ProtocolsInfo struct {
	WebDAV     ProtocolToggle `json:"webdav"`
	FTP        ProtocolToggle `json:"ftp"`
	Connection ConnectionInfo `json:"connection"`
}

type ConnectionInfo struct {
	WebDAVURL          string `json:"webdav_url"`
	FTPHost            string `json:"ftp_host"`
	FTPPort            int    `json:"ftp_port"`
	FTPUsername        string `json:"ftp_username"`
	WebDAVUsernameHint string `json:"webdav_username_hint"`
}

func (s *Service) GetProtocols(claims *auth.Claims, volumeID uuid.UUID, host string, ftpPort int) (*ProtocolsInfo, error) {
	vol, err := s.Get(claims, volumeID)
	if err != nil {
		return nil, err
	}
	return s.protocolsInfo(vol, host, ftpPort), nil
}

func (s *Service) UpdateProtocols(claims *auth.Claims, volumeID uuid.UUID, input UpdateProtocolsInput) (*ProtocolsInfo, error) {
	if input.WebDAV == nil && input.FTP == nil {
		return nil, ErrInvalidFilter
	}

	vol, err := s.findVolume(volumeID)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeVolume(claims, vol); err != nil {
		return nil, err
	}

	cfg, err := ReadVolumeConfig(vol.RootPath)
	if err != nil {
		return nil, err
	}
	if cfg.Protocols == (ProtocolsConfig{}) {
		cfg.Protocols = DefaultProtocolsConfig()
	}

	if input.WebDAV != nil {
		cfg.Protocols.WebDAV = *input.WebDAV
		vol.Protocols.WebDAV = *input.WebDAV
	}
	if input.FTP != nil {
		cfg.Protocols.FTP = *input.FTP
		vol.Protocols.FTP = *input.FTP
	}

	if err := WriteVolumeConfig(vol.RootPath, cfg); err != nil {
		return nil, err
	}
	vol.UpdatedAt = vol.UpdatedAt.UTC()
	if err := s.db.Save(vol).Error; err != nil {
		return nil, err
	}
	if s.events != nil {
		s.events.VolumeUpdated(context.Background(), VolumeUpdatedEvent{Volume: vol})
	}
	return s.protocolsInfo(vol, "", 0), nil
}

func (s *Service) protocolsInfo(vol *Volume, host string, ftpPort int) *ProtocolsInfo {
	scheme := "http"
	if host == "" {
		host = "localhost"
	}
	webdavURL := scheme + "://" + host + "/dav/volumes/" + vol.ID.String() + "/"
	ftpHost := host
	if ftpPort == 0 {
		ftpPort = 2121
	}
	return &ProtocolsInfo{
		WebDAV: vol.Protocols.WebDAV,
		FTP:    vol.Protocols.FTP,
		Connection: ConnectionInfo{
			WebDAVURL:          webdavURL,
			FTPHost:            ftpHost,
			FTPPort:            ftpPort,
			FTPUsername:        vol.ID.String(),
			WebDAVUsernameHint: "your lcloud email",
		},
	}
}

func (s *Service) AuthorizeVolume(claims *auth.Claims, vol *Volume) error {
	return s.authorizeVolume(claims, vol)
}

func (s *Service) ProtocolEnabled(vol *Volume, kind ProtocolKind, instanceWebDAV, instanceFTP bool) bool {
	switch kind {
	case ProtocolWebDAV:
		if !instanceWebDAV {
			return false
		}
		return vol.Protocols.WebDAV.Enabled
	case ProtocolFTP:
		if !instanceFTP {
			return false
		}
		return vol.Protocols.FTP.Enabled
	default:
		return false
	}
}

func NormalizeProtocols(cfg ProtocolsConfig) ProtocolsConfig {
	if cfg == (ProtocolsConfig{}) {
		return DefaultProtocolsConfig()
	}
	return cfg
}

var ErrProtocolsDisabled = errors.New("protocol access disabled")
