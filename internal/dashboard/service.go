package dashboard

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/theo-henon/lcloud/internal/auth"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) GetLayout(_ context.Context, claims *auth.Claims) (DashboardResponse, error) {
	catalog := CatalogForRole(claims.Role)

	var row UserDashboardLayout
	err := s.db.First(&row, "user_id = ?", claims.UserID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return DashboardResponse{
			Layout:    DefaultLayout(claims.Role),
			Catalog:   catalog,
			UpdatedAt: nil,
		}, nil
	}
	if err != nil {
		return DashboardResponse{}, err
	}

	updatedAt := row.UpdatedAt
	return DashboardResponse{
		Layout:    row.Layout,
		Catalog:   catalog,
		UpdatedAt: &updatedAt,
	}, nil
}

func (s *Service) SaveLayout(_ context.Context, claims *auth.Claims, layout Layout) (DashboardResponse, error) {
	if err := ValidateLayout(layout, claims.Role); err != nil {
		return DashboardResponse{}, err
	}

	now := time.Now().UTC()
	row := UserDashboardLayout{
		UserID:    claims.UserID,
		Layout:    layout,
		UpdatedAt: now,
	}

	if err := s.db.Save(&row).Error; err != nil {
		return DashboardResponse{}, err
	}

	return DashboardResponse{
		Layout:    row.Layout,
		Catalog:   CatalogForRole(claims.Role),
		UpdatedAt: &now,
	}, nil
}

func (s *Service) ResetLayout(_ context.Context, userID uuid.UUID) error {
	result := s.db.Delete(&UserDashboardLayout{}, "user_id = ?", userID)
	return result.Error
}
