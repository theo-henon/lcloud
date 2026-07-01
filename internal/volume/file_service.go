package volume

import (
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

type FileService struct {
	volumes      *Service
	paths        *PathResolver
	metadata     *MetadataCache
	thumbnails   *ThumbnailGenerator
	indexManager *indexer.IndexManager
}

func NewFileService(volumes *Service, indexManager *indexer.IndexManager) *FileService {
	return &FileService{
		volumes:      volumes,
		paths:        NewPathResolver(),
		metadata:     NewMetadataCache(),
		thumbnails:   NewThumbnailGenerator(),
		indexManager: indexManager,
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
	if err := QuotaAllows(vol, size); err != nil {
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

	hasher := sha256.New()
	tee := io.TeeReader(reader, hasher)
	file, err := os.Create(absPath)
	if err != nil {
		return nil, err
	}
	written, err := io.Copy(file, tee)
	if closeErr := file.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(absPath)
		return nil, err
	}
	if size > 0 && written != size {
		_ = os.Remove(absPath)
		return nil, ErrQuotaExceeded
	}

	mimeType := detectMime(absPath, filename)
	modified := time.Now().UTC()
	if info, err := os.Stat(absPath); err == nil {
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
		imageFile, err := os.Open(absPath)
		if err == nil {
			if thumbPath, err := s.thumbnails.Generate(vol.RootPath, fileID, imageFile); err == nil {
				record.ThumbnailPath = thumbPath
			}
			_ = imageFile.Close()
		}
	}

	if err := s.metadata.Write(vol.RootPath, record); err != nil {
		return nil, err
	}

	idx, err := s.indexManager.Get(vol.RootPath)
	if err != nil {
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
		return nil, err
	}

	if err := s.volumes.syncUsage(vol, vol.UsedBytes+written); err != nil {
		return nil, err
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

	absPath, err := s.paths.ResolveUserdata(vol.RootPath, relPath)
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

	record, metaErr := s.metadata.ReadByRelativePath(vol.RootPath, filepath.ToSlash(cleanRelativePath(relPath)))
	if metaErr == nil {
		_ = s.thumbnails.Delete(vol.RootPath, record.ID)
		_ = s.metadata.Delete(vol.RootPath, record.ID)
		if idx, err := s.indexManager.Get(vol.RootPath); err == nil {
			_ = idx.Delete(record.RelativePath)
		}
	}

	if err := os.Remove(absPath); err != nil {
		return err
	}
	return s.volumes.syncUsage(vol, vol.UsedBytes-info.Size())
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
