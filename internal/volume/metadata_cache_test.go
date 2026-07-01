package volume

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMetadataCache_ListAll_Empty(t *testing.T) {
	root := t.TempDir()
	cache := NewMetadataCache()

	records, err := cache.ListAll(root)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("expected 0 records, got %d", len(records))
	}
}

func TestMetadataCache_ListAll_ExcludesStatsFile(t *testing.T) {
	root := t.TempDir()
	cache := NewMetadataCache()

	statsPath := filepath.Join(root, CacheDir, MetadataDir, StatsCacheFile)
	if err := os.MkdirAll(filepath.Dir(statsPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(statsPath, []byte(`{"computed_at":"2026-01-01T00:00:00Z"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	record := FileMetadataRecord{
		ID:           "file-1",
		Name:         "photo.jpg",
		RelativePath: "photo.jpg",
		MimeType:     "image/jpeg",
		SizeBytes:    1024,
		ModifiedAt:   time.Now().UTC(),
	}
	if err := cache.Write(root, record); err != nil {
		t.Fatal(err)
	}

	records, err := cache.ListAll(root)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].Name != "photo.jpg" {
		t.Fatalf("expected photo.jpg, got %s", records[0].Name)
	}
}

func TestMetadataCache_ListAll_MultipleFiles(t *testing.T) {
	root := t.TempDir()
	cache := NewMetadataCache()

	for i, name := range []string{"a.jpg", "b.png", "c.pdf"} {
		record := FileMetadataRecord{
			ID:           filepath.Base(name),
			Name:         name,
			RelativePath: name,
			MimeType:     "application/octet-stream",
			SizeBytes:    int64(i + 1) * 100,
			ModifiedAt:   time.Now().UTC(),
		}
		if err := cache.Write(root, record); err != nil {
			t.Fatal(err)
		}
	}

	records, err := cache.ListAll(root)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("expected 3 records, got %d", len(records))
	}
}
