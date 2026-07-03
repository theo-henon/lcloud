package settings

import (
	"github.com/theo-henon/lcloud/internal/auth"
	"gorm.io/gorm"
)

const singletonID uint = 1

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) EnsureDefaults() error {
	var count int64
	if err := s.db.Model(&InstanceSettings{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return s.db.Create(&InstanceSettings{ID: singletonID, MaskDiskNames: false}).Error
}

func (s *Service) Get() (InstanceSettings, error) {
	var settings InstanceSettings
	if err := s.db.First(&settings, "id = ?", singletonID).Error; err != nil {
		return InstanceSettings{}, err
	}
	return settings, nil
}

func (s *Service) Public() (PublicSettings, error) {
	settings, err := s.Get()
	if err != nil {
		return PublicSettings{}, err
	}
	return s.toPublic(settings), nil
}

func (s *Service) Update(input UpdateSettingsInput) (PublicSettings, error) {
	settings, err := s.Get()
	if err != nil {
		return PublicSettings{}, err
	}
	if input.MaskDiskNames != nil {
		settings.MaskDiskNames = *input.MaskDiskNames
	}
	if input.ProtocolsWebDAVEnabled != nil {
		settings.ProtocolsWebDAVEnabled = *input.ProtocolsWebDAVEnabled
	}
	if input.ProtocolsFTPEnabled != nil {
		settings.ProtocolsFTPEnabled = *input.ProtocolsFTPEnabled
	}
	if err := s.db.Save(&settings).Error; err != nil {
		return PublicSettings{}, err
	}
	return s.toPublic(settings), nil
}

func (s *Service) ProtocolsWebDAVEnabled() (bool, error) {
	settings, err := s.Get()
	if err != nil {
		return false, err
	}
	return settings.ProtocolsWebDAVEnabled, nil
}

func (s *Service) ProtocolsFTPEnabled() (bool, error) {
	settings, err := s.Get()
	if err != nil {
		return false, err
	}
	return settings.ProtocolsFTPEnabled, nil
}

func (s *Service) toPublic(settings InstanceSettings) PublicSettings {
	return PublicSettings{
		MaskDiskNames:          settings.MaskDiskNames,
		ProtocolsWebDAVEnabled: settings.ProtocolsWebDAVEnabled,
		ProtocolsFTPEnabled:    settings.ProtocolsFTPEnabled,
	}
}

func (s *Service) ShouldMaskFor(_ *auth.Claims) (bool, error) {
	settings, err := s.Get()
	if err != nil {
		return false, err
	}
	return settings.MaskDiskNames, nil
}
