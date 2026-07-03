package notification

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/settings"
	"github.com/theo-henon/lcloud/internal/task"
	"github.com/theo-henon/lcloud/internal/volume"
	"gorm.io/gorm"
)

type Service struct {
	db       *gorm.DB
	auth     *auth.Service
	volumes  *volume.Service
	tasks    *task.Service
	settings *settings.Service
}

func NewService(
	db *gorm.DB,
	authService *auth.Service,
	volumes *volume.Service,
	tasks *task.Service,
	settingsService *settings.Service,
) *Service {
	return &Service{
		db:       db,
		auth:     authService,
		volumes:  volumes,
		tasks:    tasks,
		settings: settingsService,
	}
}

type createInput struct {
	UserID        uuid.UUID
	Type          string
	Title         string
	Body          string
	LinkPath      string
	VolumeID      *uuid.UUID
	TaskID        *uuid.UUID
	SourceEventID *string
}

func (s *Service) List(_ context.Context, userID uuid.UUID, limit, offset int) ([]Response, int64, error) {
	if limit <= 0 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	base := s.db.Model(&Notification{}).
		Where("user_id = ? AND dismissed_at IS NULL", userID)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []Notification
	if err := base.Order("created_at DESC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	out := make([]Response, len(rows))
	for i, row := range rows {
		out[i] = ToResponse(row)
	}
	return out, total, nil
}

func (s *Service) UnreadCount(_ context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := s.db.Model(&Notification{}).
		Where("user_id = ? AND dismissed_at IS NULL AND read_at IS NULL", userID).
		Count(&count).Error
	return count, err
}

func (s *Service) MarkRead(_ context.Context, userID, id uuid.UUID) (Response, error) {
	var row Notification
	if err := s.db.First(&row, "id = ? AND user_id = ?", id, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Response{}, ErrNotFound
		}
		return Response{}, err
	}
	if row.ReadAt == nil {
		now := time.Now().UTC()
		row.ReadAt = &now
		if err := s.db.Save(&row).Error; err != nil {
			return Response{}, err
		}
	}
	return ToResponse(row), nil
}

func (s *Service) MarkAllRead(_ context.Context, userID uuid.UUID) (int64, error) {
	now := time.Now().UTC()
	result := s.db.Model(&Notification{}).
		Where("user_id = ? AND dismissed_at IS NULL AND read_at IS NULL", userID).
		Update("read_at", now)
	return result.RowsAffected, result.Error
}

func (s *Service) Dismiss(_ context.Context, userID, id uuid.UUID) error {
	var row Notification
	if err := s.db.First(&row, "id = ? AND user_id = ?", id, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	if row.DismissedAt != nil {
		return nil
	}
	now := time.Now().UTC()
	row.DismissedAt = &now
	return s.db.Save(&row).Error
}

func (s *Service) create(ctx context.Context, input createInput) error {
	if input.SourceEventID != nil && *input.SourceEventID != "" {
		var existing int64
		if err := s.db.Model(&Notification{}).
			Where("user_id = ? AND type = ? AND source_event_id = ?", input.UserID, input.Type, *input.SourceEventID).
			Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			return nil
		}
	}

	row := Notification{
		ID:            uuid.New(),
		UserID:        input.UserID,
		Type:          input.Type,
		Title:         input.Title,
		Body:          input.Body,
		LinkPath:      input.LinkPath,
		VolumeID:      input.VolumeID,
		TaskID:        input.TaskID,
		SourceEventID: input.SourceEventID,
		CreatedAt:     time.Now().UTC(),
	}
	if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
		return err
	}
	return s.trimRetention(input.UserID)
}

// Seed inserts a notification for integration tests.
func (s *Service) Seed(ctx context.Context, userID uuid.UUID, title, sourceEventID string) (Response, error) {
	source := sourceEventID
	if err := s.create(ctx, createInput{
		UserID:        userID,
		Type:          TypeVolumeUsageAlert,
		Title:         title,
		Body:          "test body",
		LinkPath:      "/volumes",
		SourceEventID: &source,
	}); err != nil {
		return Response{}, err
	}
	items, _, err := s.List(ctx, userID, 1, 0)
	if err != nil || len(items) == 0 {
		return Response{}, err
	}
	return items[0], nil
}

func (s *Service) trimRetention(userID uuid.UUID) error {
	var ids []uuid.UUID
	err := s.db.Model(&Notification{}).
		Select("id").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(maxNotificationsPerUser).
		Pluck("id", &ids).Error
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	return s.db.Delete(&Notification{}, "id IN ?", ids).Error
}

func (s *Service) NotifyDeletionRequest(ctx context.Context, req volume.VolumeDeletionRequest) error {
	enabled, err := s.settings.NotifyAdminsOnDeletionRequest()
	if err != nil {
		return err
	}
	if !enabled {
		return nil
	}

	adminIDs, err := s.auth.ListActiveAdminIDs()
	if err != nil {
		return err
	}
	if len(adminIDs) == 0 {
		return nil
	}

	vol, err := s.volumes.GetByID(req.VolumeID)
	if err != nil {
		return nil
	}

	emails, err := s.auth.EmailsByIDs([]uuid.UUID{req.UserID})
	if err != nil {
		return err
	}
	userEmail := emails[req.UserID]
	if userEmail == "" {
		userEmail = "A user"
	}

	volumeID := req.VolumeID
	sourceID := req.ID.String()
	title := fmt.Sprintf("Deletion requested: %s", vol.Name)
	body := fmt.Sprintf("%s requested deletion of volume « %s ».", userEmail, vol.Name)
	linkPath := "/volumes"

	for _, adminID := range adminIDs {
		eventID := sourceID
		if err := s.create(ctx, createInput{
			UserID:        adminID,
			Type:          TypeVolumeDeletionRequest,
			Title:         title,
			Body:          body,
			LinkPath:      linkPath,
			VolumeID:      &volumeID,
			SourceEventID: &eventID,
		}); err != nil {
			return err
		}
	}
	return nil
}
