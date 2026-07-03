package dashboard

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

const LayoutVersion = 1

type WidgetPlacement struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
	W    int    `json:"w"`
	H    int    `json:"h"`
}

type Layout struct {
	Version int               `json:"version"`
	Widgets []WidgetPlacement `json:"widgets"`
}

func (l Layout) Value() (driver.Value, error) {
	return json.Marshal(l)
}

func (l *Layout) Scan(value any) error {
	if value == nil {
		*l = Layout{Version: LayoutVersion, Widgets: []WidgetPlacement{}}
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, l)
	case string:
		return json.Unmarshal([]byte(v), l)
	default:
		return errors.New("unsupported layout scan type")
	}
}

type CatalogEntry struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	DefaultW    int    `json:"default_w"`
	DefaultH    int    `json:"default_h"`
	AdminOnly   bool   `json:"admin_only"`
}

type DashboardResponse struct {
	Layout    Layout         `json:"layout"`
	Catalog   []CatalogEntry `json:"catalog"`
	UpdatedAt *time.Time     `json:"updated_at"`
}

type UserDashboardLayout struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	Layout    Layout    `gorm:"type:jsonb;not null" json:"layout"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (UserDashboardLayout) TableName() string {
	return "user_dashboard_layouts"
}

type PatchLayoutInput struct {
	Layout Layout `json:"layout" binding:"required"`
}
