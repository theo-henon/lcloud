package volume

import (
	"context"
	"os"
	"path/filepath"

	"github.com/theo-henon/lcloud/internal/indexer"
)

type MacroOps struct {
	volumes          *Service
	paths            *PathResolver
	metadata         *MetadataCache
	thumbnails       *ThumbnailGenerator
	indexManager     *indexer.IndexManager
	statsInvalidator StatsInvalidator
	events           EventPublisher
}

func NewMacroOps(
	volumes *Service,
	indexManager *indexer.IndexManager,
	statsInvalidator StatsInvalidator,
	events EventPublisher,
) *MacroOps {
	return &MacroOps{
		volumes:          volumes,
		paths:            NewPathResolver(),
		metadata:         NewMetadataCache(),
		thumbnails:       NewThumbnailGenerator(),
		indexManager:     indexManager,
		statsInvalidator: statsInvalidator,
		events:           events,
	}
}

func (m *MacroOps) DeleteFileInternal(ctx context.Context, vol *Volume, relPath string) error {
	_ = ctx
	relPath = filepath.ToSlash(cleanRelativePath(relPath))

	absPath, err := m.paths.ResolveUserdata(vol.RootPath, relPath)
	if err != nil {
		return err
	}
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotFound
		}
		return err
	}
	if info.IsDir() {
		return ErrNotDirectory
	}

	record, metaErr := m.metadata.ReadByRelativePath(vol.RootPath, relPath)
	if metaErr == nil {
		_ = m.thumbnails.Delete(vol.RootPath, record.ID)
		_ = m.metadata.Delete(vol.RootPath, record.ID)
		if idx, err := m.indexManager.Get(vol.RootPath); err == nil {
			_ = idx.Delete(record.RelativePath)
		}
	}

	if err := os.Remove(absPath); err != nil {
		return err
	}
	if err := m.volumes.syncUsage(vol, vol.UsedBytes-info.Size()); err != nil {
		return err
	}
	m.invalidateStats(vol.RootPath)

	if m.events != nil {
		m.events.FileDeleted(context.Background(), FileDeletedEvent{
			VolumeID:     vol.ID,
			RelativePath: relPath,
			SizeBytes:    info.Size(),
		})
	}
	return nil
}

func (m *MacroOps) MoveFileInternal(ctx context.Context, vol *Volume, fromRel, toRel string) error {
	return MoveFile(ctx, m.fileOpDeps(), vol, fromRel, toRel)
}

func (m *MacroOps) InvalidateStats(rootPath string) {
	m.invalidateStats(rootPath)
}

func (m *MacroOps) invalidateStats(rootPath string) {
	if m.statsInvalidator != nil {
		_ = m.statsInvalidator.Invalidate(rootPath)
	}
}

func (m *MacroOps) Metadata() *MetadataCache {
	return m.metadata
}

func (m *MacroOps) Paths() *PathResolver {
	return m.paths
}

func (m *MacroOps) Thumbnails() *ThumbnailGenerator {
	return m.thumbnails
}

func (m *MacroOps) IndexManager() *indexer.IndexManager {
	return m.indexManager
}
