package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var manifestIDPattern = regexp.MustCompile(`^[a-z0-9-]+$`)

type Registry struct {
	db          *gorm.DB
	pluginsPath string
	logger      *slog.Logger
}

func NewRegistry(db *gorm.DB, pluginsPath string, logger *slog.Logger) *Registry {
	if logger == nil {
		logger = slog.Default()
	}
	return &Registry{db: db, pluginsPath: pluginsPath, logger: logger}
}

func (r *Registry) ScanAndLoad() ([]Plugin, error) {
	if err := os.MkdirAll(r.pluginsPath, 0o755); err != nil {
		return nil, fmt.Errorf("create plugins path: %w", err)
	}

	entries, err := os.ReadDir(r.pluginsPath)
	if err != nil {
		return nil, fmt.Errorf("read plugins path: %w", err)
	}

	seen := make(map[string]struct{})
	var loaded []Plugin
	now := time.Now().UTC()

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".json") {
			continue
		}

		binaryPath := filepath.Join(r.pluginsPath, name)
		info, err := entry.Info()
		if err != nil {
			r.logger.Warn("plugin registry: stat failed", "path", binaryPath, "error", err)
			continue
		}
		if info.Mode()&0o111 == 0 {
			r.logger.Warn("plugin registry: skipping non-executable", "path", binaryPath)
			continue
		}

		manifestPath := binaryPath + ".json"
		manifest, err := parseManifest(manifestPath)
		if err != nil {
			r.logger.Warn("plugin registry: skipping binary without valid manifest", "path", binaryPath, "error", err)
			continue
		}
		if err := validateManifest(manifest); err != nil {
			r.logger.Warn("plugin registry: invalid manifest", "path", manifestPath, "error", err)
			continue
		}
		if _, ok := seen[manifest.ID]; ok {
			r.logger.Warn("plugin registry: duplicate plugin id", "id", manifest.ID)
			continue
		}
		seen[manifest.ID] = struct{}{}

		record := Plugin{
			ID:            manifest.ID,
			Name:          manifest.Name,
			Version:       manifest.Version,
			Description:   manifest.Description,
			Author:        manifest.Author,
			BinaryPath:    binaryPath,
			ManifestPath:  manifestPath,
			Subscriptions: StringArray(manifest.Subscribe),
			DeclaredEmits: StringArray(manifest.Emit),
			DiscoveredAt:  now,
			UpdatedAt:     now,
		}

		var existing Plugin
		err = r.db.First(&existing, "id = ?", record.ID).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			record.Enabled = true
			record.Status = PluginStatusStopped
			if err := r.db.Create(&record).Error; err != nil {
				return nil, err
			}
		case err != nil:
			return nil, err
		default:
			record.Enabled = existing.Enabled
			record.Status = PluginStatusStopped
			record.DiscoveredAt = existing.DiscoveredAt
			if err := r.db.Model(&existing).Updates(map[string]any{
				"name":           record.Name,
				"version":        record.Version,
				"description":    record.Description,
				"author":         record.Author,
				"binary_path":    record.BinaryPath,
				"manifest_path":  record.ManifestPath,
				"subscriptions":  record.Subscriptions,
				"declared_emits": record.DeclaredEmits,
				"status":         PluginStatusStopped,
				"updated_at":     now,
			}).Error; err != nil {
				return nil, err
			}
			record = existing
			record.Name = manifest.Name
			record.Version = manifest.Version
			record.Description = manifest.Description
			record.Author = manifest.Author
			record.BinaryPath = binaryPath
			record.ManifestPath = manifestPath
			record.Subscriptions = StringArray(manifest.Subscribe)
			record.DeclaredEmits = StringArray(manifest.Emit)
			record.Status = PluginStatusStopped
		}

		loaded = append(loaded, record)
	}

	return loaded, nil
}

func parseManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}
	return &manifest, nil
}

func validateManifest(m *Manifest) error {
	if m.ID == "" || !manifestIDPattern.MatchString(m.ID) {
		return errors.New("invalid manifest id")
	}
	if strings.TrimSpace(m.Name) == "" {
		return errors.New("manifest name is required")
	}
	if strings.TrimSpace(m.Version) == "" {
		return errors.New("manifest version is required")
	}
	return nil
}

func (r *Registry) List() ([]Plugin, error) {
	var plugins []Plugin
	if err := r.db.Order("name asc").Find(&plugins).Error; err != nil {
		return nil, err
	}
	return plugins, nil
}

func (r *Registry) Get(id string) (*Plugin, error) {
	var record Plugin
	if err := r.db.First(&record, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPluginNotFound
		}
		return nil, err
	}
	return &record, nil
}

func (r *Registry) SetEnabled(id string, enabled bool) error {
	result := r.db.Model(&Plugin{}).Where("id = ?", id).Updates(map[string]any{
		"enabled":    enabled,
		"updated_at": time.Now().UTC(),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrPluginNotFound
	}
	return nil
}

func (r *Registry) SetStatus(id string, status PluginStatus, lastError string) error {
	return r.db.Model(&Plugin{}).Where("id = ?", id).Updates(map[string]any{
		"status":     status,
		"last_error": lastError,
		"updated_at": time.Now().UTC(),
	}).Error
}

func (r *Registry) AppendLog(ctx context.Context, req LogRequest) (*PluginLogEntry, error) {
	entry := PluginLogEntry{
		ID:        uuid.New(),
		PluginID:  req.PluginID,
		EventType: req.EventType,
		Level:     string(req.Level),
		Message:   req.Message,
		CreatedAt: time.Now().UTC(),
	}
	if req.VolumeID != "" {
		volID, err := uuid.Parse(req.VolumeID)
		if err == nil {
			entry.VolumeID = &volID
		}
	}
	if req.Payload != nil {
		entry.Payload = JSON(req.Payload)
	}
	if err := r.db.WithContext(ctx).Create(&entry).Error; err != nil {
		return nil, err
	}
	if err := r.purgeOldLogs(ctx, req.PluginID); err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *Registry) purgeOldLogs(ctx context.Context, pluginID string) error {
	var ids []uuid.UUID
	if err := r.db.WithContext(ctx).Model(&PluginLogEntry{}).
		Where("plugin_id = ?", pluginID).
		Order("created_at desc").
		Offset(maxLogEntriesPerPlugin).
		Pluck("id", &ids).Error; err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Delete(&PluginLogEntry{}, "id IN ?", ids).Error
}

type LogFilter struct {
	PluginID  string
	VolumeID  *uuid.UUID
	Limit     int
	OwnerID   *uuid.UUID
	IsAdmin   bool
}

func (r *Registry) ListLogs(ctx context.Context, filter LogFilter) ([]PluginLogEntry, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	query := r.db.WithContext(ctx).Model(&PluginLogEntry{}).Order("created_at desc").Limit(limit)
	if filter.PluginID != "" {
		query = query.Where("plugin_id = ?", filter.PluginID)
	}
	if filter.VolumeID != nil {
		query = query.Where("volume_id = ?", *filter.VolumeID)
	}
	if !filter.IsAdmin && filter.OwnerID != nil {
		query = query.Where(`volume_id IN (SELECT id FROM volumes WHERE owner_id = ?)`, *filter.OwnerID)
	}

	var entries []PluginLogEntry
	if err := query.Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}
