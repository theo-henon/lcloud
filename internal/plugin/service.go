package plugin

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/volume"
	"gorm.io/gorm"
)

type Service struct {
	registry *Registry
	runtime  *Runtime
	bus      *EventBus
	bridge   *EventBridge
	logger   *slog.Logger
}

func NewService(db *gorm.DB, pluginsPath string, volumes *volume.Service, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	bus := NewEventBus()
	registry := NewRegistry(db, pluginsPath, logger)
	volumeProvider := newVolumeProvider(volumes)
	runtime := NewRuntime(registry, bus, volumeProvider, logger)
	runtime.Subscribe()

	return &Service{
		registry: registry,
		runtime:  runtime,
		bus:      bus,
		bridge:   NewEventBridge(bus),
		logger:   logger,
	}
}

func (s *Service) Publisher() volume.EventPublisher {
	return s.bridge
}

func (s *Service) Start(ctx context.Context) error {
	if _, err := s.registry.ScanAndLoad(); err != nil {
		return err
	}
	return s.runtime.StartEnabled(ctx)
}

func (s *Service) Stop() {
	s.runtime.StopAll()
}

func (s *Service) List() ([]Plugin, error) {
	return s.registry.List()
}

func (s *Service) Get(id string) (*Plugin, error) {
	return s.registry.Get(id)
}

func (s *Service) SetEnabled(ctx context.Context, id string, enabled bool) error {
	if err := s.registry.SetEnabled(id, enabled); err != nil {
		return err
	}
	if enabled {
		return s.runtime.Start(ctx, id)
	}
	s.runtime.Stop(id)
	return nil
}

func (s *Service) ListLogs(ctx context.Context, claims *auth.Claims, filter LogFilter) ([]PluginLogEntry, error) {
	filter.IsAdmin = claims.Role == auth.RoleAdmin
	if !filter.IsAdmin {
		filter.OwnerID = &claims.UserID
	}
	return s.registry.ListLogs(ctx, filter)
}

func (s *Service) PublishEvent(ctx context.Context, event Event) {
	s.bus.Publish(ctx, event)
}

type volumeProvider struct {
	volumes *volume.Service
}

func newVolumeProvider(volumes *volume.Service) VolumeProvider {
	return &volumeProvider{volumes: volumes}
}

func (p *volumeProvider) GetVolumeConfig(ctx context.Context, volumeID uuid.UUID) (*VolumeConfigView, error) {
	if p.volumes == nil {
		return nil, volume.ErrVolumeNotFound
	}
	vol, err := p.volumes.GetByID(volumeID)
	if err != nil {
		return nil, err
	}
	return &VolumeConfigView{
		ID:         vol.ID.String(),
		Name:       vol.Name,
		QuotaBytes: vol.QuotaBytes,
		Filters:    vol.Filters,
	}, nil
}

func (p *volumeProvider) VolumeRootPath(ctx context.Context, volumeID uuid.UUID) (string, error) {
	if p.volumes == nil {
		return "", volume.ErrVolumeNotFound
	}
	vol, err := p.volumes.GetByID(volumeID)
	if err != nil {
		return "", err
	}
	return vol.RootPath, nil
}
