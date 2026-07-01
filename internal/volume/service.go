package volume

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/indexer"
	"gorm.io/gorm"
)

type CreateVolumeInput struct {
	Name       string
	DiskPath   string
	QuotaBytes int64
	Filters    Filters
}

type PatchVolumeInput struct {
	Name       *string
	QuotaBytes *int64
	Filters    *Filters
}

type Service struct {
	db           *gorm.DB
	disks        *DiskRegistry
	indexManager *indexer.IndexManager
}

func NewService(db *gorm.DB, disks *DiskRegistry, indexManager *indexer.IndexManager) *Service {
	return &Service{
		db:           db,
		disks:        disks,
		indexManager: indexManager,
	}
}

func (s *Service) List(claims *auth.Claims) ([]Volume, error) {
	query := s.db.Order("created_at desc")
	if claims.Role != auth.RoleAdmin {
		query = query.Where("owner_id = ?", claims.UserID)
	}

	var volumes []Volume
	if err := query.Find(&volumes).Error; err != nil {
		return nil, err
	}
	return volumes, nil
}

func (s *Service) Get(claims *auth.Claims, id uuid.UUID) (*Volume, error) {
	vol, err := s.findVolume(id)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeVolume(claims, vol); err != nil {
		return nil, err
	}
	return vol, nil
}

func (s *Service) Create(claims *auth.Claims, input CreateVolumeInput) (*Volume, error) {
	if !s.disks.IsRegistered(input.DiskPath) {
		return nil, ErrDiskNotFound
	}

	filters := NormalizeFilters(input.Filters)
	if err := ValidateFilters(filters); err != nil {
		return nil, err
	}

	var existing int64
	if err := s.db.Model(&Volume{}).
		Where("owner_id = ? AND name = ?", claims.UserID, input.Name).
		Count(&existing).Error; err != nil {
		return nil, err
	}
	if existing > 0 {
		return nil, ErrVolumeNameTaken
	}

	id := uuid.New()
	rootPath := filepath.Join(input.DiskPath, id.String())
	if err := createVolumeLayout(rootPath); err != nil {
		return nil, fmt.Errorf("create volume layout: %w", err)
	}

	now := time.Now().UTC()
	cfg := VolumeConfig{
		ID:         id,
		Name:       input.Name,
		OwnerID:    claims.UserID,
		QuotaBytes: input.QuotaBytes,
		Filters:    filters,
		Encryption: EncryptionConfig{Enabled: false, Method: nil},
		CreatedAt:  now,
		DiskPath:   input.DiskPath,
		UsedBytes:  0,
	}
	if err := WriteVolumeConfig(rootPath, cfg); err != nil {
		_ = removeVolumeTree(rootPath)
		return nil, err
	}

	vol := volumeFromConfig(cfg, rootPath)
	if err := s.db.Create(vol).Error; err != nil {
		_ = removeVolumeTree(rootPath)
		return nil, err
	}
	return vol, nil
}

func (s *Service) Patch(claims *auth.Claims, id uuid.UUID, input PatchVolumeInput) (*Volume, error) {
	vol, err := s.findVolume(id)
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

	if input.Name != nil {
		var existing int64
		if err := s.db.Model(&Volume{}).
			Where("owner_id = ? AND name = ? AND id <> ?", vol.OwnerID, *input.Name, vol.ID).
			Count(&existing).Error; err != nil {
			return nil, err
		}
		if existing > 0 {
			return nil, ErrVolumeNameTaken
		}
		cfg.Name = *input.Name
		vol.Name = *input.Name
	}
	if input.QuotaBytes != nil {
		cfg.QuotaBytes = *input.QuotaBytes
		vol.QuotaBytes = *input.QuotaBytes
	}
	if input.Filters != nil {
		normalized := NormalizeFilters(*input.Filters)
		if err := ValidateFilters(normalized); err != nil {
			return nil, err
		}
		cfg.Filters = normalized
		vol.Filters = normalized
	}

	if err := WriteVolumeConfig(vol.RootPath, cfg); err != nil {
		return nil, err
	}
	vol.UpdatedAt = time.Now().UTC()
	if err := s.db.Save(vol).Error; err != nil {
		return nil, err
	}
	return vol, nil
}

func (s *Service) Delete(claims *auth.Claims, id uuid.UUID, force bool) error {
	vol, err := s.findVolume(id)
	if err != nil {
		return err
	}
	if err := s.authorizeVolume(claims, vol); err != nil {
		return err
	}

	empty, err := isUserdataEmpty(vol.RootPath)
	if err != nil {
		return err
	}
	if !empty {
		if claims.Role != auth.RoleAdmin || !force {
			return ErrVolumeNotEmpty
		}
	}

	if err := s.indexManager.Close(vol.RootPath); err != nil {
		return err
	}
	if err := removeVolumeTree(vol.RootPath); err != nil {
		return err
	}
	return s.db.Delete(&Volume{}, "id = ?", vol.ID).Error
}

func (s *Service) findVolume(id uuid.UUID) (*Volume, error) {
	var vol Volume
	if err := s.db.First(&vol, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVolumeNotFound
		}
		return nil, err
	}
	return &vol, nil
}

func (s *Service) authorizeVolume(claims *auth.Claims, vol *Volume) error {
	if claims.Role == auth.RoleAdmin || vol.OwnerID == claims.UserID {
		return nil
	}
	return ErrForbidden
}

func (s *Service) syncUsage(vol *Volume, usedBytes int64) error {
	cfg, err := ReadVolumeConfig(vol.RootPath)
	if err != nil {
		return err
	}
	cfg.UsedBytes = usedBytes
	vol.UsedBytes = usedBytes
	if err := WriteVolumeConfig(vol.RootPath, cfg); err != nil {
		return err
	}
	return s.db.Model(vol).Update("used_bytes", usedBytes).Error
}
