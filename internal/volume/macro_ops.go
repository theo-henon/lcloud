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
	_ = ctx
	fromRel = filepath.ToSlash(cleanRelativePath(fromRel))
	toRel = filepath.ToSlash(cleanRelativePath(toRel))

	fromAbs, err := m.paths.ResolveUserdata(vol.RootPath, fromRel)
	if err != nil {
		return err
	}
	toAbs, err := m.paths.ResolveUserdata(vol.RootPath, toRel)
	if err != nil {
		return err
	}

	fromInfo, err := os.Stat(fromAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotFound
		}
		return err
	}
	if fromInfo.IsDir() {
		return ErrNotDirectory
	}
	if _, err := os.Stat(toAbs); err == nil {
		return ErrDirectoryExists
	}

	if err := os.MkdirAll(filepath.Dir(toAbs), 0o755); err != nil {
		return err
	}
	if err := os.Rename(fromAbs, toAbs); err != nil {
		return err
	}

	record, metaErr := m.metadata.ReadByRelativePath(vol.RootPath, fromRel)
	if metaErr == nil {
		record.RelativePath = toRel
		record.Name = filepath.Base(toRel)
		_ = m.metadata.Write(vol.RootPath, record)
		if idx, err := m.indexManager.Get(vol.RootPath); err == nil {
			_ = idx.Delete(fromRel)
			_ = idx.Index(indexer.FileMetadata{
				ID:           record.ID,
				Name:         record.Name,
				RelativePath: record.RelativePath,
				MimeType:     record.MimeType,
				SizeBytes:    record.SizeBytes,
				ModifiedAt:   record.ModifiedAt,
				SHA256:       record.SHA256,
			})
		}
	}

	m.invalidateStats(vol.RootPath)

	if m.events != nil {
		m.events.FileMoved(context.Background(), FileMovedEvent{
			VolumeID:     vol.ID,
			FromPath:     fromRel,
			ToPath:       toRel,
			SizeBytes:    fromInfo.Size(),
		})
	}
	return nil
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
