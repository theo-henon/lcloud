package indexer

import "time"

type FileMetadata struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	RelativePath string    `json:"relative_path"`
	MimeType     string    `json:"mime_type"`
	SizeBytes    int64     `json:"size_bytes"`
	ModifiedAt   time.Time `json:"modified_at"`
	SHA256       string    `json:"sha256"`
}

type SearchQuery struct {
	Term string
}

type VolumeIndexer interface {
	Index(file FileMetadata) error
	Delete(path string) error
	Search(query SearchQuery) ([]FileMetadata, error)
	Rebuild(volumePath string) error
	Close() error
}
