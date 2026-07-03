package ftp

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"sync"

	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/protocols"
)

type Server struct {
	gateway  *protocols.Gateway
	driver   *mainDriver
	ftp      *ftpserver.FtpServer
	mu       sync.Mutex
	running  bool
}

type mainDriver struct {
	gateway      *protocols.Gateway
	listenAddr   string
	pasvAddress  string
	pasvMin      int
	pasvMax      int
}

func NewServer(gateway *protocols.Gateway, listenAddr, pasvAddress string, pasvMin, pasvMax int) *Server {
	return &Server{
		gateway: gateway,
		driver: &mainDriver{
			gateway:     gateway,
			listenAddr:  listenAddr,
			pasvAddress: pasvAddress,
			pasvMin:     pasvMin,
			pasvMax:     pasvMax,
		},
	}
}

func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return nil
	}

	enabled, err := s.gateway.Settings().ProtocolsFTPEnabled()
	if err != nil {
		return err
	}
	if !enabled {
		return nil
	}

	s.ftp = ftpserver.NewFtpServer(s.driver)
	if err := s.ftp.Listen(); err != nil {
		return fmt.Errorf("ftp listen: %w", err)
	}

	go func() {
		if err := s.ftp.Serve(); err != nil {
			log.Printf("ftp server stopped: %v", err)
		}
	}()

	s.running = true
	return nil
}

func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running || s.ftp == nil {
		return
	}
	_ = s.ftp.Stop()
	s.running = false
}

func (d *mainDriver) GetSettings() (*ftpserver.Settings, error) {
	return &ftpserver.Settings{
		ListenAddr:               d.listenAddr,
		PublicHost:               d.pasvAddress,
		Banner:                   "lcloud FTP",
		PassiveTransferPortRange: ftpserver.PortRange{Start: d.pasvMin, End: d.pasvMax},
		DisableActiveMode:        false,
	}, nil
}

func (d *mainDriver) ClientConnected(cc ftpserver.ClientContext) (string, error) {
	return "220 lcloud FTP ready", nil
}

func (d *mainDriver) ClientDisconnected(cc ftpserver.ClientContext) {}

func (d *mainDriver) AuthUser(cc ftpserver.ClientContext, user, pass string) (ftpserver.ClientDriver, error) {
	claims, volumeID, err := d.gateway.FTPLogin(cc.RemoteAddr().String(), user, pass)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, errors.New("530 invalid credentials")
		}
		if errors.Is(err, auth.ErrUserDisabled) {
			return nil, errors.New("530 account disabled")
		}
		return nil, errors.New("530 access denied")
	}
	cc.SetPath("/")
	return newVolumeFs(d.gateway, claims, volumeID), nil
}

func (d *mainDriver) GetTLSConfig() (*tls.Config, error) {
	return nil, errors.New("TLS not supported")
}
