package volume

import (
	"context"
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
	DiskID     string
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
	events       EventPublisher
}

func NewService(db *gorm.DB, disks *DiskRegistry, indexManager *indexer.IndexManager) *Service {
	return &Service{
		db:           db,
		disks:        disks,
		indexManager: indexManager,
	}
}

func (s *Service) SetEventPublisher(events EventPublisher) {
	s.events = events
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

func (s *Service) GetByID(id uuid.UUID) (*Volume, error) {
	return s.findVolume(id)
}

func (s *Service) ListAll() ([]Volume, error) {
	var volumes []Volume
	if err := s.db.Order("created_at asc").Find(&volumes).Error; err != nil {
		return nil, err
	}
	return volumes, nil
}

func (s *Service) Create(claims *auth.Claims, input CreateVolumeInput) (*Volume, error) {
	diskPath, err := s.resolveCreateDiskPath(input.DiskPath, input.DiskID)
	if err != nil {
		return nil, err
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
	rootPath := filepath.Join(diskPath, id.String())
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
		DiskPath:   diskPath,
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
	if s.events != nil {
		s.events.VolumeCreated(context.Background(), VolumeCreatedEvent{Volume: vol})
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
	if s.events != nil {
		s.events.VolumeUpdated(context.Background(), VolumeUpdatedEvent{Volume: vol})
	}
	return vol, nil
}

func (s *Service) Delete(claims *auth.Claims, id uuid.UUID, force bool) error {
	if claims.Role != auth.RoleAdmin {
		return ErrForbidden
	}

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
	if err := s.db.Delete(&Volume{}, "id = ?", vol.ID).Error; err != nil {
		return err
	}
	if s.events != nil {
		s.events.VolumeDeleted(context.Background(), VolumeDeletedEvent{
			VolumeID: vol.ID,
			Name:     vol.Name,
		})
	}
	_ = s.dismissDeletionRequestsForVolume(vol.ID)
	return nil
}

func (s *Service) resolveCreateDiskPath(diskPath, diskID string) (string, error) {
	if diskPath != "" {
		if s.disks.IsRegistered(diskPath) {
			return diskPath, nil
		}
		return "", ErrDiskNotFound
	}
	if diskID != "" {
		resolved, ok := s.disks.ResolvePathByID(diskID)
		if !ok || !s.disks.IsRegistered(resolved) {
			return "", ErrDiskNotFound
		}
		return resolved, nil
	}
	return "", ErrDiskNotFound
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

func (s *Service) RequestDeletion(claims *auth.Claims, volumeID uuid.UUID) error {
	vol, err := s.findVolume(volumeID)
	if err != nil {
		return err
	}
	if vol.OwnerID != claims.UserID {
		return ErrForbidden
	}

	var existing int64
	if err := s.db.Model(&VolumeDeletionRequest{}).
		Where("volume_id = ? AND status = ?", volumeID, DeletionRequestPending).
		Count(&existing).Error; err != nil {
		return err
	}
	if existing > 0 {
		return ErrDeletionRequestExists
	}

	req := &VolumeDeletionRequest{
		ID:       uuid.New(),
		VolumeID: volumeID,
		UserID:   claims.UserID,
		Status:   DeletionRequestPending,
	}
	return s.db.Create(req).Error
}

func (s *Service) ListPendingDeletionRequests() ([]DeletionRequestResponse, error) {
	var results []DeletionRequestResponse
	err := s.db.Table("volume_deletion_requests AS r").
		Select("r.id, r.volume_id, v.name AS volume_name, r.user_id, u.email AS user_email, r.created_at").
		Joins("JOIN volumes v ON v.id = r.volume_id").
		Joins("JOIN users u ON u.id = r.user_id").
		Where("r.status = ?", DeletionRequestPending).
		Order("r.created_at ASC").
		Scan(&results).Error
	return results, err
}

func (s *Service) DismissDeletionRequest(id uuid.UUID) error {
	var req VolumeDeletionRequest
	if err := s.db.First(&req, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDeletionRequestNotFound
		}
		return err
	}
	if req.Status != DeletionRequestPending {
		return ErrDeletionRequestNotFound
	}
	now := time.Now()
	req.Status = DeletionRequestDismissed
	req.DismissedAt = &now
	return s.db.Save(&req).Error
}

func (s *Service) PendingDeletionVolumeIDs(userID uuid.UUID) (map[uuid.UUID]bool, error) {
	var rows []VolumeDeletionRequest
	if err := s.db.Where("user_id = ? AND status = ?", userID, DeletionRequestPending).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]bool, len(rows))
	for _, row := range rows {
		out[row.VolumeID] = true
	}
	return out, nil
}

func (s *Service) OwnerEmailsByIDs(ids []uuid.UUID) (map[uuid.UUID]string, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]string{}, nil
	}

	unique := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		unique[id] = struct{}{}
	}
	deduped := make([]uuid.UUID, 0, len(unique))
	for id := range unique {
		deduped = append(deduped, id)
	}

	var rows []struct {
		ID    uuid.UUID
		Email string
	}
	if err := s.db.Table("users").Select("id, email").Where("id IN ?", deduped).Scan(&rows).Error; err != nil {
		return nil, err
	}

	out := make(map[uuid.UUID]string, len(rows))
	for _, row := range rows {
		out[row.ID] = row.Email
	}
	return out, nil
}

func (s *Service) dismissDeletionRequestsForVolume(volumeID uuid.UUID) error {
	now := time.Now()
	return s.db.Model(&VolumeDeletionRequest{}).
		Where("volume_id = ? AND status = ?", volumeID, DeletionRequestPending).
		Updates(map[string]any{
			"status":       DeletionRequestDismissed,
			"dismissed_at": now,
		}).Error
}
