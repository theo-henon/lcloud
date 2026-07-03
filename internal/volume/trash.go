package volume

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/indexer"
)

const TrashMetaFile = "meta.json"

type TrashMeta struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	OriginalPath string    `json:"original_path"`
	MimeType     string    `json:"mime_type"`
	SizeBytes    int64     `json:"size_bytes"`
	SHA256       string    `json:"sha256,omitempty"`
	DeletedAt    time.Time `json:"deleted_at"`
	ModifiedAt   time.Time `json:"modified_at"`
}

type TrashEntry struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	OriginalPath string    `json:"original_path"`
	MimeType     string    `json:"mime_type"`
	SizeBytes    int64     `json:"size_bytes"`
	DeletedAt    time.Time `json:"deleted_at"`
	ModifiedAt   time.Time `json:"modified_at"`
	HasThumbnail bool      `json:"has_thumbnail,omitempty"`
}

type TrashListing struct {
	Items []TrashEntry `json:"items"`
	Total int          `json:"total"`
}

type TrashOpDeps struct {
	FileOpDeps
	thumbnails *ThumbnailGenerator
}

type TrashService struct {
	volumes          *Service
	paths            *PathResolver
	metadata         *MetadataCache
	thumbnails       *ThumbnailGenerator
	indexManager     *indexer.IndexManager
	statsInvalidator StatsInvalidator
	events           EventPublisher
}

func NewTrashService(volumes *Service, indexManager *indexer.IndexManager, statsInvalidator StatsInvalidator) *TrashService {
	return &TrashService{
		volumes:          volumes,
		paths:            NewPathResolver(),
		metadata:         NewMetadataCache(),
		thumbnails:       NewThumbnailGenerator(),
		indexManager:     indexManager,
		statsInvalidator: statsInvalidator,
	}
}

func (s *TrashService) SetEventPublisher(events EventPublisher) {
	s.events = events
}

func (s *TrashService) deps() TrashOpDeps {
	return TrashOpDeps{
		FileOpDeps: FileOpDeps{
			volumes:          s.volumes,
			paths:            s.paths,
			metadata:         s.metadata,
			indexManager:     s.indexManager,
			statsInvalidator: s.statsInvalidator,
			events:           s.events,
		},
		thumbnails: s.thumbnails,
	}
}

func (m *MacroOps) trashOpDeps() TrashOpDeps {
	return TrashOpDeps{
		FileOpDeps: m.fileOpDeps(),
		thumbnails: m.thumbnails,
	}
}

func (s *FileService) trashOpDeps() TrashOpDeps {
	return TrashOpDeps{
		FileOpDeps: s.fileOpDeps(),
		thumbnails: s.thumbnails,
	}
}

func trashRoot(rootPath string) string {
	return filepath.Join(rootPath, TrashDir)
}

func trashItemDir(rootPath, fileID string) string {
	return filepath.Join(trashRoot(rootPath), fileID)
}

func validateTrashFileID(fileID string) error {
	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return ErrTrashItemNotFound
	}
	if fileID == "." || fileID == ".." {
		return ErrPathTraversal
	}
	if strings.ContainsAny(fileID, `/\`) || strings.Contains(fileID, "..") {
		return ErrPathTraversal
	}
	return nil
}

func resolveTrashItemDir(rootPath, fileID string) (string, error) {
	if err := validateTrashFileID(fileID); err != nil {
		return "", err
	}
	trashAbs, err := filepath.Abs(trashRoot(rootPath))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrPathTraversal, err)
	}
	itemDir, err := filepath.Abs(filepath.Join(trashAbs, fileID))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrPathTraversal, err)
	}
	if !pathWithinRoot(trashAbs, itemDir) {
		return "", ErrPathTraversal
	}
	return itemDir, nil
}

func ensureTrashDir(rootPath string) error {
	return os.MkdirAll(trashRoot(rootPath), 0o755)
}

func readTrashMeta(itemDir string) (TrashMeta, error) {
	data, err := os.ReadFile(filepath.Join(itemDir, TrashMetaFile))
	if err != nil {
		return TrashMeta{}, err
	}
	var meta TrashMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return TrashMeta{}, err
	}
	return meta, nil
}

func writeTrashMeta(itemDir string, meta TrashMeta) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(itemDir, TrashMetaFile), data, 0o644)
}

func TrashFileInternal(ctx context.Context, deps TrashOpDeps, vol *Volume, relPath string) error {
	_ = ctx
	relPath = filepath.ToSlash(cleanRelativePath(relPath))

	absPath, err := deps.paths.ResolveUserdata(vol.RootPath, relPath)
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

	record, metaErr := deps.metadata.ReadByRelativePath(vol.RootPath, relPath)
	fileID := uuid.NewString()
	name := filepath.Base(relPath)
	mimeType := ""
	sha256sum := ""
	modifiedAt := info.ModTime().UTC()
	sizeBytes := info.Size()

	if metaErr == nil {
		fileID = record.ID
		name = record.Name
		mimeType = record.MimeType
		sha256sum = record.SHA256
		if !record.ModifiedAt.IsZero() {
			modifiedAt = record.ModifiedAt
		}
		sizeBytes = record.SizeBytes
		if sizeBytes == 0 {
			sizeBytes = info.Size()
		}
	} else if mimeType == "" {
		mimeType = detectMime(absPath, name)
	}

	deletedAt := time.Now().UTC()
	if err := ensureTrashDir(vol.RootPath); err != nil {
		return err
	}

	itemDir := trashItemDir(vol.RootPath, fileID)
	if err := os.MkdirAll(itemDir, 0o755); err != nil {
		return err
	}

	meta := TrashMeta{
		ID:           fileID,
		Name:         name,
		OriginalPath: relPath,
		MimeType:     mimeType,
		SizeBytes:    sizeBytes,
		SHA256:       sha256sum,
		DeletedAt:    deletedAt,
		ModifiedAt:   modifiedAt,
	}
	if err := writeTrashMeta(itemDir, meta); err != nil {
		_ = os.RemoveAll(itemDir)
		return err
	}

	destBlob := filepath.Join(itemDir, name)
	if err := os.Rename(absPath, destBlob); err != nil {
		_ = os.RemoveAll(itemDir)
		return err
	}

	if metaErr == nil {
		_ = deps.metadata.Delete(vol.RootPath, record.ID)
		if idx, err := deps.indexManager.Get(vol.RootPath); err == nil {
			_ = idx.Delete(record.RelativePath)
		}
	}

	invalidateFileOpStats(deps.FileOpDeps, vol.RootPath)

	if deps.events != nil {
		deps.events.FileTrashed(context.Background(), FileTrashedEvent{
			VolumeID:     vol.ID,
			FileID:       fileID,
			OriginalPath: relPath,
			SizeBytes:    sizeBytes,
			DeletedAt:    deletedAt,
		})
	}
	return nil
}

func PurgeTrashItemInternal(ctx context.Context, deps TrashOpDeps, vol *Volume, fileID string) error {
	_ = ctx
	itemDir, err := resolveTrashItemDir(vol.RootPath, fileID)
	if err != nil {
		return err
	}
	meta, err := readTrashMeta(itemDir)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrTrashItemNotFound
		}
		return err
	}

	_ = deps.thumbnails.Delete(vol.RootPath, fileID)
	if err := os.RemoveAll(itemDir); err != nil {
		return err
	}

	newUsed := vol.UsedBytes - meta.SizeBytes
	if newUsed < 0 {
		newUsed = 0
	}
	if err := deps.volumes.syncUsage(vol, newUsed); err != nil {
		return err
	}
	invalidateFileOpStats(deps.FileOpDeps, vol.RootPath)

	if deps.events != nil {
		deps.events.FileDeleted(context.Background(), FileDeletedEvent{
			VolumeID:     vol.ID,
			RelativePath: meta.OriginalPath,
			SizeBytes:    meta.SizeBytes,
		})
	}
	return nil
}

func listAllTrashMeta(rootPath string) ([]TrashMeta, error) {
	dir := trashRoot(rootPath)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	items := make([]TrashMeta, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		meta, err := readTrashMeta(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}
		items = append(items, meta)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].DeletedAt.After(items[j].DeletedAt)
	})
	return items, nil
}

func trashEntryFromMeta(rootPath string, meta TrashMeta, thumbs *ThumbnailGenerator) TrashEntry {
	hasThumb := false
	if thumbs != nil {
		if _, err := os.Stat(thumbs.ThumbnailPath(rootPath, meta.ID)); err == nil {
			hasThumb = true
		}
	}
	return TrashEntry{
		ID:           meta.ID,
		Name:         meta.Name,
		OriginalPath: meta.OriginalPath,
		MimeType:     meta.MimeType,
		SizeBytes:    meta.SizeBytes,
		DeletedAt:    meta.DeletedAt,
		ModifiedAt:   meta.ModifiedAt,
		HasThumbnail: hasThumb,
	}
}

func RestoreTrashItemInternal(ctx context.Context, deps TrashOpDeps, vol *Volume, fileID string) (*FileEntry, error) {
	_ = ctx
	itemDir, err := resolveTrashItemDir(vol.RootPath, fileID)
	if err != nil {
		return nil, err
	}
	meta, err := readTrashMeta(itemDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrTrashItemNotFound
		}
		return nil, err
	}

	targetRel := filepath.ToSlash(cleanRelativePath(meta.OriginalPath))
	targetAbs, err := deps.paths.ResolveUserdata(vol.RootPath, targetRel)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(targetAbs); err == nil {
		return nil, ErrPathOccupied
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	srcBlob := filepath.Join(itemDir, meta.Name)
	if err := os.MkdirAll(filepath.Dir(targetAbs), 0o755); err != nil {
		return nil, err
	}
	if err := os.Rename(srcBlob, targetAbs); err != nil {
		return nil, err
	}

	thumbPath := ""
	if _, err := os.Stat(deps.thumbnails.ThumbnailPath(vol.RootPath, meta.ID)); err == nil {
		thumbPath = deps.thumbnails.ThumbnailPath(vol.RootPath, meta.ID)
	}

	record := FileMetadataRecord{
		ID:            meta.ID,
		Name:          meta.Name,
		RelativePath:  targetRel,
		MimeType:      meta.MimeType,
		SizeBytes:     meta.SizeBytes,
		SHA256:        meta.SHA256,
		ModifiedAt:    meta.ModifiedAt,
		ThumbnailPath: thumbPath,
	}
	if err := deps.metadata.Write(vol.RootPath, record); err != nil {
		return nil, err
	}
	if idx, err := deps.indexManager.Get(vol.RootPath); err == nil {
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

	_ = os.RemoveAll(itemDir)
	invalidateFileOpStats(deps.FileOpDeps, vol.RootPath)

	if deps.events != nil {
		deps.events.FileRestored(context.Background(), FileRestoredEvent{
			VolumeID:     vol.ID,
			FileID:       meta.ID,
			RestoredPath: targetRel,
			SizeBytes:    meta.SizeBytes,
		})
	}

	return entryForPath(vol.RootPath, targetRel, false)
}

func PurgeTrashOlderThan(ctx context.Context, deps TrashOpDeps, vol *Volume, days int, dryRun bool) (int, []string, error) {
	cutoff := time.Now().UTC().AddDate(0, 0, -days)
	items, err := listAllTrashMeta(vol.RootPath)
	if err != nil {
		return 0, nil, err
	}

	affected := 0
	preview := make([]string, 0, 20)
	for _, meta := range items {
		if err := ctx.Err(); err != nil {
			return affected, preview, err
		}
		if !meta.DeletedAt.Before(cutoff) {
			continue
		}
		if dryRun {
			affected++
			if len(preview) < 20 {
				preview = append(preview, meta.OriginalPath)
			}
			continue
		}
		if err := PurgeTrashItemInternal(ctx, deps, vol, meta.ID); err != nil {
			if err == ErrTrashItemNotFound {
				continue
			}
			return affected, preview, err
		}
		affected++
	}
	return affected, preview, nil
}

func (m *MacroOps) TrashFileInternal(ctx context.Context, vol *Volume, relPath string) error {
	return TrashFileInternal(ctx, m.trashOpDeps(), vol, relPath)
}

func (m *MacroOps) PurgeTrashItemInternal(ctx context.Context, vol *Volume, fileID string) error {
	return PurgeTrashItemInternal(ctx, m.trashOpDeps(), vol, fileID)
}

func (m *MacroOps) PurgeTrashOlderThan(ctx context.Context, vol *Volume, days int, dryRun bool) (int, []string, error) {
	return PurgeTrashOlderThan(ctx, m.trashOpDeps(), vol, days, dryRun)
}

func (s *TrashService) List(claims *auth.Claims, volumeID uuid.UUID, limit, offset int) (*TrashListing, error) {
	vol, err := s.volumes.Get(claims, volumeID)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	metas, err := listAllTrashMeta(vol.RootPath)
	if err != nil {
		return nil, err
	}

	total := len(metas)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}

	items := make([]TrashEntry, 0, end-offset)
	for _, meta := range metas[offset:end] {
		items = append(items, trashEntryFromMeta(vol.RootPath, meta, s.thumbnails))
	}
	return &TrashListing{Items: items, Total: total}, nil
}

func (s *TrashService) Restore(claims *auth.Claims, volumeID uuid.UUID, fileID string) (*FileEntry, error) {
	vol, err := s.volumes.Get(claims, volumeID)
	if err != nil {
		return nil, err
	}
	return RestoreTrashItemInternal(context.Background(), s.deps(), vol, fileID)
}

func (s *TrashService) PurgeOne(claims *auth.Claims, volumeID uuid.UUID, fileID string) error {
	vol, err := s.volumes.Get(claims, volumeID)
	if err != nil {
		return err
	}
	return PurgeTrashItemInternal(context.Background(), s.deps(), vol, fileID)
}

func (s *TrashService) Empty(claims *auth.Claims, volumeID uuid.UUID) (int, error) {
	vol, err := s.volumes.Get(claims, volumeID)
	if err != nil {
		return 0, err
	}
	metas, err := listAllTrashMeta(vol.RootPath)
	if err != nil {
		return 0, err
	}
	purged := 0
	for _, meta := range metas {
		if err := PurgeTrashItemInternal(context.Background(), s.deps(), vol, meta.ID); err != nil {
			if err == ErrTrashItemNotFound {
				continue
			}
			return purged, err
		}
		purged++
	}
	return purged, nil
}

func (s *TrashService) PurgeOlderThanAllVolumes(ctx context.Context, days int) (int, error) {
	volumes, err := s.volumes.ListAll()
	if err != nil {
		return 0, err
	}
	total := 0
	for i := range volumes {
		count, _, err := PurgeTrashOlderThan(ctx, s.deps(), &volumes[i], days, false)
		if err != nil {
			return total, err
		}
		total += count
	}
	return total, nil
}

func (s *TrashService) OpenThumbnail(claims *auth.Claims, volumeID uuid.UUID, fileID string) (*os.File, error) {
	vol, err := s.volumes.Get(claims, volumeID)
	if err != nil {
		return nil, err
	}
	if err := validateTrashFileID(fileID); err != nil {
		return nil, err
	}
	if _, err := os.Stat(s.thumbnails.ThumbnailPath(vol.RootPath, fileID)); os.IsNotExist(err) {
		return nil, ErrFileNotFound
	}
	return s.thumbnails.Open(vol.RootPath, fileID)
}
