package volume

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type FileMetadataRecord struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	RelativePath  string    `json:"relative_path"`
	MimeType      string    `json:"mime_type"`
	SizeBytes     int64     `json:"size_bytes"`
	SHA256        string    `json:"sha256"`
	ModifiedAt    time.Time `json:"modified_at"`
	ThumbnailPath string    `json:"thumbnail_path,omitempty"`
}

const StatsCacheFile = "stats.json"

type MetadataCache struct{}

func NewMetadataCache() *MetadataCache {
	return &MetadataCache{}
}

func (c *MetadataCache) metadataDir(rootPath string) string {
	return filepath.Join(rootPath, CacheDir, MetadataDir)
}

func (c *MetadataCache) metadataPath(rootPath, fileID string) string {
	return filepath.Join(c.metadataDir(rootPath), fileID+".json")
}

func (c *MetadataCache) Write(rootPath string, record FileMetadataRecord) error {
	if err := os.MkdirAll(c.metadataDir(rootPath), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(c.metadataPath(rootPath, record.ID), data, 0o644)
}

func (c *MetadataCache) Read(rootPath, fileID string) (FileMetadataRecord, error) {
	data, err := os.ReadFile(c.metadataPath(rootPath, fileID))
	if err != nil {
		return FileMetadataRecord{}, err
	}
	var record FileMetadataRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return FileMetadataRecord{}, err
	}
	return record, nil
}

func (c *MetadataCache) Delete(rootPath, fileID string) error {
	err := os.Remove(c.metadataPath(rootPath, fileID))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (c *MetadataCache) ReadByRelativePath(rootPath, relPath string) (FileMetadataRecord, error) {
	entries, err := os.ReadDir(c.metadataDir(rootPath))
	if err != nil {
		return FileMetadataRecord{}, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(c.metadataDir(rootPath), entry.Name()))
		if err != nil {
			continue
		}
		var record FileMetadataRecord
		if err := json.Unmarshal(data, &record); err != nil {
			continue
		}
		if record.RelativePath == relPath {
			return record, nil
		}
	}
	return FileMetadataRecord{}, os.ErrNotExist
}

func (c *MetadataCache) DeleteByRelativePath(rootPath, relPath string) error {
	record, err := c.ReadByRelativePath(rootPath, relPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return c.Delete(rootPath, record.ID)
}

func (c *MetadataCache) ListAll(rootPath string) ([]FileMetadataRecord, error) {
	dir := c.metadataDir(rootPath)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []FileMetadataRecord{}, nil
		}
		return nil, err
	}

	records := make([]FileMetadataRecord, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == StatsCacheFile {
			continue
		}
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}
		var record FileMetadataRecord
		if err := json.Unmarshal(data, &record); err != nil {
			continue
		}
		records = append(records, record)
	}
	return records, nil
}
