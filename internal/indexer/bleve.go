package indexer

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
)

type BleveIndexer struct {
	indexPath string
	index     bleve.Index
	mu        sync.Mutex
}

func OpenBleveIndexer(volumeRoot string) (*BleveIndexer, error) {
	indexPath := filepath.Join(volumeRoot, "cache", "index")
	cacheDir := filepath.Join(volumeRoot, "cache")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("create cache dir: %w", err)
	}

	index, err := bleve.Open(indexPath)
	if err == bleve.ErrorIndexPathDoesNotExist {
		mapping := bleve.NewIndexMapping()
		mapping.DefaultMapping = newFileMapping()
		index, err = bleve.New(indexPath, mapping)
	}
	if err != nil {
		return nil, fmt.Errorf("open bleve index: %w", err)
	}

	return &BleveIndexer{indexPath: indexPath, index: index}, nil
}

func newFileMapping() *mapping.DocumentMapping {
	doc := bleve.NewDocumentMapping()
	for _, field := range []string{"name", "relative_path", "mime_type", "sha256"} {
		fieldMapping := bleve.NewTextFieldMapping()
		fieldMapping.Store = true
		doc.AddFieldMappingsAt(field, fieldMapping)
	}
	for _, field := range []string{"size_bytes"} {
		fieldMapping := bleve.NewNumericFieldMapping()
		fieldMapping.Store = true
		doc.AddFieldMappingsAt(field, fieldMapping)
	}
	modifiedMapping := bleve.NewDateTimeFieldMapping()
	modifiedMapping.Store = true
	doc.AddFieldMappingsAt("modified_at", modifiedMapping)
	return doc
}

func (b *BleveIndexer) Index(file FileMetadata) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	doc := map[string]any{
		"name":          file.Name,
		"relative_path": file.RelativePath,
		"mime_type":     file.MimeType,
		"size_bytes":    file.SizeBytes,
		"modified_at":   file.ModifiedAt,
		"sha256":        file.SHA256,
	}
	return b.index.Index(file.RelativePath, doc)
}

func (b *BleveIndexer) Delete(path string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.index.Delete(path)
}

func (b *BleveIndexer) Search(query SearchQuery) ([]FileMetadata, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	search := bleve.NewMatchQuery(query.Term)
	searchRequest := bleve.NewSearchRequest(search)
	searchRequest.Fields = []string{"*"}
	result, err := b.index.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	files := make([]FileMetadata, 0, len(result.Hits))
	for _, hit := range result.Hits {
		files = append(files, fieldsToMetadata(hit.ID, hit.Fields))
	}
	return files, nil
}

func (b *BleveIndexer) Rebuild(volumePath string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err := b.index.Close(); err != nil {
		return err
	}
	if err := os.RemoveAll(b.indexPath); err != nil {
		return err
	}

	index, err := bleve.New(b.indexPath, func() *mapping.IndexMappingImpl {
		m := bleve.NewIndexMapping()
		m.DefaultMapping = newFileMapping()
		return m
	}())
	if err != nil {
		return err
	}
	b.index = index
	return nil
}

func (b *BleveIndexer) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.index == nil {
		return nil
	}
	err := b.index.Close()
	b.index = nil
	return err
}

func fieldsToMetadata(id string, fields map[string]any) FileMetadata {
	meta := FileMetadata{RelativePath: id}
	if value, ok := fields["name"].(string); ok {
		meta.Name = value
	}
	if value, ok := fields["mime_type"].(string); ok {
		meta.MimeType = value
	}
	if value, ok := fields["sha256"].(string); ok {
		meta.SHA256 = value
	}
	if value, ok := fields["size_bytes"].(float64); ok {
		meta.SizeBytes = int64(value)
	}
	return meta
}

type IndexManager struct {
	mu      sync.Mutex
	indexes map[string]*BleveIndexer
}

func NewIndexManager() *IndexManager {
	return &IndexManager{indexes: make(map[string]*BleveIndexer)}
}

func (m *IndexManager) Get(volumeRoot string) (*BleveIndexer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if idx, ok := m.indexes[volumeRoot]; ok {
		return idx, nil
	}

	idx, err := OpenBleveIndexer(volumeRoot)
	if err != nil {
		return nil, err
	}
	m.indexes[volumeRoot] = idx
	return idx, nil
}

func (m *IndexManager) Close(volumeRoot string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	idx, ok := m.indexes[volumeRoot]
	if !ok {
		return nil
	}
	delete(m.indexes, volumeRoot)
	return idx.Close()
}

func (m *IndexManager) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var firstErr error
	for root, idx := range m.indexes {
		if err := idx.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		delete(m.indexes, root)
	}
	return firstErr
}
