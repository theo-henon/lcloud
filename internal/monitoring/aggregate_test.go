package monitoring

import (
	"testing"
	"time"

	"github.com/theo-henon/lcloud/internal/volume"
)

func TestAggregateRecords(t *testing.T) {
	records := []volume.FileMetadataRecord{
		{MimeType: "image/jpeg", SizeBytes: 800},
		{MimeType: "image/png", SizeBytes: 200},
		{MimeType: "application/pdf", SizeBytes: 100},
	}

	computedAt := time.Now().UTC()
	stats := aggregateRecords(records, computedAt)

	if stats.FileCount != 3 {
		t.Fatalf("file_count = %d, want 3", stats.FileCount)
	}
	if stats.TotalBytes != 1100 {
		t.Fatalf("total_bytes = %d, want 1100", stats.TotalBytes)
	}
	if len(stats.ByCategory) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(stats.ByCategory))
	}
	if stats.ByCategory[0].Category != CategoryImages {
		t.Fatalf("top category = %s, want images", stats.ByCategory[0].Category)
	}
	if stats.ByCategory[0].Proportion < 0.9 {
		t.Fatalf("images proportion = %f, want ~0.909", stats.ByCategory[0].Proportion)
	}
}

func TestAggregateRecords_Empty(t *testing.T) {
	stats := aggregateRecords(nil, time.Now().UTC())
	if stats.FileCount != 0 || stats.TotalBytes != 0 {
		t.Fatalf("expected empty stats, got %+v", stats)
	}
}
