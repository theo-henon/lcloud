package volume

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/theo-henon/lcloud/internal/indexer"
)

type FileOpDeps struct {
	volumes          *Service
	paths            *PathResolver
	metadata         *MetadataCache
	indexManager     *indexer.IndexManager
	statsInvalidator StatsInvalidator
	events           EventPublisher
}

func (m *MacroOps) fileOpDeps() FileOpDeps {
	return FileOpDeps{
		volumes:          m.volumes,
		paths:            m.paths,
		metadata:         m.metadata,
		indexManager:     m.indexManager,
		statsInvalidator: m.statsInvalidator,
		events:           m.events,
	}
}

func (s *FileService) fileOpDeps() FileOpDeps {
	return FileOpDeps{
		volumes:          s.volumes,
		paths:            s.paths,
		metadata:         s.metadata,
		indexManager:     s.indexManager,
		statsInvalidator: s.statsInvalidator,
		events:           s.events,
	}
}

func MoveFile(ctx context.Context, deps FileOpDeps, vol *Volume, fromRel, toRel string) error {
	_ = ctx
	fromRel = filepath.ToSlash(cleanRelativePath(fromRel))
	toRel = filepath.ToSlash(cleanRelativePath(toRel))

	fromAbs, err := deps.paths.ResolveUserdata(vol.RootPath, fromRel)
	if err != nil {
		return err
	}
	toAbs, err := deps.paths.ResolveUserdata(vol.RootPath, toRel)
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
		return ErrNotAFile
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

	record, metaErr := deps.metadata.ReadByRelativePath(vol.RootPath, fromRel)
	if metaErr == nil {
		record.RelativePath = toRel
		record.Name = filepath.Base(toRel)
		_ = deps.metadata.Write(vol.RootPath, record)
		if idx, err := deps.indexManager.Get(vol.RootPath); err == nil {
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

	invalidateFileOpStats(deps, vol.RootPath)

	if deps.events != nil {
		deps.events.FileMoved(context.Background(), FileMovedEvent{
			VolumeID:  vol.ID,
			FromPath:  fromRel,
			ToPath:    toRel,
			SizeBytes: fromInfo.Size(),
		})
	}
	return nil
}

func RenameEntry(ctx context.Context, deps FileOpDeps, vol *Volume, relPath, newName string) (*FileEntry, error) {
	_ = ctx
	relPath = filepath.ToSlash(cleanRelativePath(relPath))
	if err := validateEntryName(newName); err != nil {
		return nil, err
	}

	absPath, err := deps.paths.ResolveUserdata(vol.RootPath, relPath)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}

	parent := filepath.Dir(relPath)
	if parent == "." {
		parent = ""
	}
	newRel := newName
	if parent != "" {
		newRel = filepath.ToSlash(filepath.Join(parent, newName))
	}

	newAbs, err := deps.paths.ResolveUserdata(vol.RootPath, newRel)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(newAbs); err == nil {
		return nil, ErrDirectoryExists
	}

	if err := os.Rename(absPath, newAbs); err != nil {
		return nil, err
	}

	entryType := "file"
	if info.IsDir() {
		entryType = "directory"
		if err := cascadeDirectoryRename(deps, vol, relPath, newRel); err != nil {
			return nil, err
		}
	} else if record, metaErr := deps.metadata.ReadByRelativePath(vol.RootPath, relPath); metaErr == nil {
		record.RelativePath = newRel
		record.Name = newName
		_ = deps.metadata.Write(vol.RootPath, record)
		if idx, idxErr := deps.indexManager.Get(vol.RootPath); idxErr == nil {
			_ = idx.Delete(relPath)
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

	invalidateFileOpStats(deps, vol.RootPath)

	if deps.events != nil {
		deps.events.FileRenamed(context.Background(), FileRenamedEvent{
			VolumeID:  vol.ID,
			OldPath:   relPath,
			NewPath:   newRel,
			EntryType: entryType,
		})
	}

	return entryForPath(vol.RootPath, newRel, info.IsDir())
}

func cascadeDirectoryRename(deps FileOpDeps, vol *Volume, oldRel, newRel string) error {
	records, err := deps.metadata.ListAll(vol.RootPath)
	if err != nil {
		return err
	}
	oldRel = filepath.ToSlash(oldRel)
	newRel = filepath.ToSlash(newRel)

	idx, idxErr := deps.indexManager.Get(vol.RootPath)
	for _, record := range records {
		updated, ok := replacePathPrefix(record.RelativePath, oldRel, newRel)
		if !ok {
			continue
		}
		oldPath := record.RelativePath
		record.RelativePath = updated
		record.Name = filepath.Base(updated)
		if err := deps.metadata.Write(vol.RootPath, record); err != nil {
			return err
		}
		if idxErr == nil {
			_ = idx.Delete(oldPath)
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
	return nil
}

func replacePathPrefix(path, oldPrefix, newPrefix string) (string, bool) {
	path = filepath.ToSlash(path)
	oldPrefix = filepath.ToSlash(oldPrefix)
	newPrefix = filepath.ToSlash(newPrefix)
	if path == oldPrefix {
		return newPrefix, true
	}
	prefix := oldPrefix + "/"
	if strings.HasPrefix(path, prefix) {
		return newPrefix + path[len(oldPrefix):], true
	}
	return path, false
}

func validateEntryName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" || trimmed == "." || trimmed == ".." {
		return ErrInvalidEntryName
	}
	if strings.ContainsAny(trimmed, `/\`) {
		return ErrInvalidEntryName
	}
	return nil
}

func invalidateFileOpStats(deps FileOpDeps, rootPath string) {
	if deps.statsInvalidator != nil {
		_ = deps.statsInvalidator.Invalidate(rootPath)
	}
}

func entryForPath(rootPath, relPath string, isDir bool) (*FileEntry, error) {
	relPath = filepath.ToSlash(cleanRelativePath(relPath))
	if relPath == "" {
		relPath = "."
	}
	entry := &FileEntry{
		Name: filepath.Base(relPath),
		Path: relPath,
		Type: "file",
	}
	if isDir {
		entry.Type = "directory"
		return entry, nil
	}

	absPath, err := NewPathResolver().ResolveUserdata(rootPath, relPath)
	if err != nil {
		return entry, nil
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return entry, nil
	}
	entry.SizeBytes = info.Size()
	entry.ModifiedAt = info.ModTime().UTC()

	meta := NewMetadataCache()
	if record, err := meta.ReadByRelativePath(rootPath, relPath); err == nil {
		entry.MimeType = record.MimeType
		entry.HasThumbnail = record.ThumbnailPath != ""
	}
	return entry, nil
}
