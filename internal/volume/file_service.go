package volume

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/indexer"
)

type FileEntry struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	Type         string    `json:"type"`
	SizeBytes    int64     `json:"size_bytes,omitempty"`
	MimeType     string    `json:"mime_type,omitempty"`
	ModifiedAt   time.Time `json:"modified_at,omitempty"`
	HasThumbnail bool      `json:"has_thumbnail,omitempty"`
}

type DirectoryListing struct {
	Path    string      `json:"path"`
	Entries []FileEntry `json:"entries"`
}

type StatsInvalidator interface {
	Invalidate(rootPath string) error
}

type FileService struct {
	volumes          *Service
	paths            *PathResolver
	metadata         *MetadataCache
	thumbnails       *ThumbnailGenerator
	indexManager     *indexer.IndexManager
	maxUploadBytes   int64
	statsInvalidator StatsInvalidator
	events           EventPublisher
}

func NewFileService(volumes *Service, indexManager *indexer.IndexManager, maxUploadBytes int64, statsInvalidator StatsInvalidator) *FileService {
	return &FileService{
		volumes:          volumes,
		paths:            NewPathResolver(),
		metadata:         NewMetadataCache(),
		thumbnails:       NewThumbnailGenerator(),
		indexManager:     indexManager,
		maxUploadBytes:   maxUploadBytes,
		statsInvalidator: statsInvalidator,
	}
}

func (s *FileService) SetEventPublisher(events EventPublisher) {
	s.events = events
}

func (s *FileService) invalidateStats(rootPath string) {
	if s.statsInvalidator != nil {
		_ = s.statsInvalidator.Invalidate(rootPath)
	}
}

func (s *FileService) List(claims *auth.Claims, volumeID uuid.UUID, relPath string) (*DirectoryListing, error) {
	vol, err := s.volumes.Get(claims, volumeID)
	if err != nil {
		return nil, err
	}

	absDir, err := s.paths.ResolveUserdata(vol.RootPath, relPath)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(absDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, ErrNotDirectory
	}

	entries, err := os.ReadDir(absDir)
	if err != nil {
		return nil, err
	}

	listingPath := cleanRelativePath(relPath)
	if listingPath == "" {
		listingPath = "."
	}

	result := &DirectoryListing{Path: listingPath, Entries: make([]FileEntry, 0, len(entries))}
	for _, entry := range entries {
		entryPath, err := relativeUserdataPath(vol.RootPath, filepath.Join(absDir, entry.Name()))
		if err != nil {
			continue
		}
		if entry.IsDir() {
			result.Entries = append(result.Entries, FileEntry{
				Name: entry.Name(),
				Path: entryPath,
				Type: "directory",
			})
			continue
		}

		fileInfo, err := entry.Info()
		if err != nil {
			continue
		}
		item := FileEntry{
			Name:       entry.Name(),
			Path:       entryPath,
			Type:       "file",
			SizeBytes:  fileInfo.Size(),
			ModifiedAt: fileInfo.ModTime().UTC(),
		}
		if record, err := s.metadata.ReadByRelativePath(vol.RootPath, entryPath); err == nil {
			item.MimeType = record.MimeType
			item.HasThumbnail = record.ThumbnailPath != ""
		} else {
			item.MimeType = mime.TypeByExtension(filepath.Ext(entry.Name()))
		}
		result.Entries = append(result.Entries, item)
	}
	return result, nil
}

func (s *FileService) CreateDirectory(claims *auth.Claims, volumeID uuid.UUID, relPath string) error {
	vol, err := s.volumes.Get(claims, volumeID)
	if err != nil {
		return err
	}

	clean := cleanRelativePath(relPath)
	if clean == "." || clean == "" {
		return ErrInvalidFilter
	}

	absDir, err := s.paths.ResolveUserdata(vol.RootPath, clean)
	if err != nil {
		return err
	}
	if _, err := os.Stat(absDir); err == nil {
		return ErrDirectoryExists
	}
	return os.MkdirAll(absDir, 0o755)
}

func (s *FileService) RemoveDirectory(claims *auth.Claims, volumeID uuid.UUID, relPath string) error {
	vol, err := s.volumes.Get(claims, volumeID)
	if err != nil {
		return err
	}

	clean := cleanRelativePath(relPath)
	if clean == "." || clean == "" {
		return ErrInvalidFilter
	}

	absDir, err := s.paths.ResolveUserdata(vol.RootPath, clean)
	if err != nil {
		return err
	}
	info, err := os.Stat(absDir)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotFound
		}
		return err
	}
	if !info.IsDir() {
		return ErrNotDirectory
	}

	entries, err := os.ReadDir(absDir)
	if err != nil {
		return err
	}
	if len(entries) > 0 {
		return ErrDirectoryNotEmpty
	}
	return os.Remove(absDir)
}

func (s *FileService) Upload(claims *auth.Claims, volumeID uuid.UUID, relDir, filename string, reader io.Reader, size int64) (*FileEntry, error) {
	vol, err := s.volumes.Get(claims, volumeID)
	if err != nil {
		return nil, err
	}

	filename = filepath.Base(strings.TrimSpace(filename))
	if filename == "" || filename == "." {
		return nil, ErrFileNotFound
	}
	if err := ValidateExtension(vol.Filters, filename); err != nil {
		return nil, err
	}

	targetRel := filename
	if clean := cleanRelativePath(relDir); clean != "." && clean != "" {
		targetRel = filepath.ToSlash(filepath.Join(clean, filename))
	}

	absPath, err := s.paths.ResolveUserdata(vol.RootPath, targetRel)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return nil, err
	}

	var previousSize int64
	var previousRecord *FileMetadataRecord
	if info, err := os.Stat(absPath); err == nil {
		if info.IsDir() {
			return nil, ErrNotDirectory
		}
		previousSize = info.Size()
		if rec, err := s.metadata.ReadByRelativePath(vol.RootPath, targetRel); err == nil {
			previousRecord = &rec
		}
	}

	if size > 0 {
		if err := QuotaAllows(vol, size-previousSize); err != nil {
			return nil, err
		}
	}

	tmpPath := absPath + ".upload-" + uuid.NewString()
	promoted := false
	defer func() {
		if !promoted {
			_ = os.Remove(tmpPath)
		}
	}()

	hasher := sha256.New()
	limited := io.LimitReader(reader, s.maxUploadBytes+1)
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return nil, err
	}
	written, err := io.Copy(tmpFile, io.TeeReader(limited, hasher))
	if closeErr := tmpFile.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	if err != nil {
		return nil, err
	}
	if written > s.maxUploadBytes {
		return nil, ErrUploadTooLarge
	}
	if size > 0 && written != size {
		return nil, ErrUploadSizeMismatch
	}
	if err := QuotaAllows(vol, written-previousSize); err != nil {
		return nil, err
	}

	mimeType := detectMime(tmpPath, filename)
	modified := time.Now().UTC()
	if info, err := os.Stat(tmpPath); err == nil {
		modified = info.ModTime().UTC()
	}

	fileID := uuid.NewString()
	record := FileMetadataRecord{
		ID:           fileID,
		Name:         filename,
		RelativePath: targetRel,
		MimeType:     mimeType,
		SizeBytes:    written,
		SHA256:       hex.EncodeToString(hasher.Sum(nil)),
		ModifiedAt:   modified,
	}

	if isImageMime(mimeType) {
		imageFile, err := os.Open(tmpPath)
		if err == nil {
			if thumbPath, err := s.thumbnails.Generate(vol.RootPath, fileID, imageFile); err == nil {
				record.ThumbnailPath = thumbPath
			}
			_ = imageFile.Close()
		}
	}

	if err := os.Rename(tmpPath, absPath); err != nil {
		_ = s.thumbnails.Delete(vol.RootPath, fileID)
		return nil, err
	}
	promoted = true

	if previousRecord != nil {
		_ = s.thumbnails.Delete(vol.RootPath, previousRecord.ID)
		_ = s.metadata.Delete(vol.RootPath, previousRecord.ID)
	}

	if err := s.metadata.Write(vol.RootPath, record); err != nil {
		s.rollbackUploadedFile(vol, absPath, record)
		return nil, err
	}

	idx, err := s.indexManager.Get(vol.RootPath)
	if err != nil {
		s.rollbackUploadedFile(vol, absPath, record)
		return nil, err
	}
	if err := idx.Index(indexer.FileMetadata{
		ID:           record.ID,
		Name:         record.Name,
		RelativePath: record.RelativePath,
		MimeType:     record.MimeType,
		SizeBytes:    record.SizeBytes,
		ModifiedAt:   record.ModifiedAt,
		SHA256:       record.SHA256,
	}); err != nil {
		s.rollbackUploadedFile(vol, absPath, record)
		return nil, err
	}

	newUsed := vol.UsedBytes - previousSize + written
	if newUsed < 0 {
		newUsed = 0
	}
	if err := s.volumes.syncUsage(vol, newUsed); err != nil {
		s.rollbackUploadedFile(vol, absPath, record)
		if idx, idxErr := s.indexManager.Get(vol.RootPath); idxErr == nil {
			_ = idx.Delete(record.RelativePath)
		}
		return nil, err
	}

	s.invalidateStats(vol.RootPath)

	if s.events != nil {
		s.events.FileUploaded(context.Background(), FileUploadedEvent{
			VolumeID:     vol.ID,
			Name:         filename,
			RelativePath: targetRel,
			MimeType:     mimeType,
			SizeBytes:    written,
			Filters:      vol.Filters,
		})
	}

	return &FileEntry{
		Name:         filename,
		Path:         targetRel,
		Type:         "file",
		SizeBytes:    written,
		MimeType:     mimeType,
		ModifiedAt:   modified,
		HasThumbnail: record.ThumbnailPath != "",
	}, nil
}

func (s *FileService) rollbackUploadedFile(vol *Volume, absPath string, record FileMetadataRecord) {
	_ = os.Remove(absPath)
	_ = s.thumbnails.Delete(vol.RootPath, record.ID)
	_ = s.metadata.Delete(vol.RootPath, record.ID)
}

func (s *FileService) Move(claims *auth.Claims, volumeID uuid.UUID, fromPath, toPath string) (*FileEntry, error) {
	vol, err := s.volumes.Get(claims, volumeID)
	if err != nil {
		return nil, err
	}
	if err := MoveFile(context.Background(), s.fileOpDeps(), vol, fromPath, toPath); err != nil {
		return nil, err
	}
	toPath = filepath.ToSlash(cleanRelativePath(toPath))
	return entryForPath(vol.RootPath, toPath, false)
}

func (s *FileService) Rename(claims *auth.Claims, volumeID uuid.UUID, relPath, newName string) (*FileEntry, error) {
	vol, err := s.volumes.Get(claims, volumeID)
	if err != nil {
		return nil, err
	}
	return RenameEntry(context.Background(), s.fileOpDeps(), vol, relPath, newName)
}

func (s *FileService) OpenContent(claims *auth.Claims, volumeID uuid.UUID, relPath string) (*os.File, string, error) {
	vol, err := s.volumes.Get(claims, volumeID)
	if err != nil {
		return nil, "", err
	}

	absPath, err := s.paths.ResolveUserdata(vol.RootPath, relPath)
	if err != nil {
		return nil, "", err
	}
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", ErrFileNotFound
		}
		return nil, "", err
	}
	if info.IsDir() {
		return nil, "", ErrNotDirectory
	}

	file, err := os.Open(absPath)
	if err != nil {
		return nil, "", err
	}
	mimeType := mime.TypeByExtension(filepath.Ext(absPath))
	if record, err := s.metadata.ReadByRelativePath(vol.RootPath, filepath.ToSlash(relPath)); err == nil {
		mimeType = record.MimeType
	}
	return file, mimeType, nil
}

func (s *FileService) OpenThumbnail(claims *auth.Claims, volumeID uuid.UUID, relPath string) (*os.File, error) {
	vol, err := s.volumes.Get(claims, volumeID)
	if err != nil {
		return nil, err
	}

	record, err := s.metadata.ReadByRelativePath(vol.RootPath, filepath.ToSlash(cleanRelativePath(relPath)))
	if err != nil {
		return nil, ErrFileNotFound
	}
	if record.ThumbnailPath == "" {
		return nil, ErrFileNotFound
	}
	return os.Open(record.ThumbnailPath)
}

func (s *FileService) Delete(claims *auth.Claims, volumeID uuid.UUID, relPath string) error {
	vol, err := s.volumes.Get(claims, volumeID)
	if err != nil {
		return err
	}
	return TrashFileInternal(context.Background(), s.trashOpDeps(), vol, relPath)
}

func (s *FileService) OpenThumbnailByID(claims *auth.Claims, volumeID uuid.UUID, fileID string) (*os.File, error) {
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

func detectMime(path, filename string) string {
	file, err := os.Open(path)
	if err != nil {
		return mime.TypeByExtension(filepath.Ext(filename))
	}
	defer file.Close()

	mtype, err := mimetype.DetectReader(file)
	if err != nil {
		return mime.TypeByExtension(filepath.Ext(filename))
	}
	return mtype.String()
}
