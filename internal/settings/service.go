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
	return PublicSettings{MaskDiskNames: settings.MaskDiskNames}, nil
}

func (s *Service) Update(input UpdateSettingsInput) (PublicSettings, error) {
	settings, err := s.Get()
	if err != nil {
		return PublicSettings{}, err
	}
	if input.MaskDiskNames != nil {
		settings.MaskDiskNames = *input.MaskDiskNames
	}
	if err := s.db.Save(&settings).Error; err != nil {
		return PublicSettings{}, err
	}
	return PublicSettings{MaskDiskNames: settings.MaskDiskNames}, nil
}

func (s *Service) ShouldMaskFor(_ *auth.Claims) (bool, error) {
	settings, err := s.Get()
	if err != nil {
		return false, err
	}
	return settings.MaskDiskNames, nil
}
