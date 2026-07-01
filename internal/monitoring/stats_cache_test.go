package monitoring

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/theo-henon/lcloud/internal/volume"
)

func TestStatsCache_RoundTrip(t *testing.T) {
	root := t.TempDir()
	cache := NewStatsCache()

	stats := cachedVolumeStats{
		ComputedAt: time.Now().UTC().Truncate(time.Second),
		FileCount:  2,
		TotalBytes: 3000,
		ByCategory: []CategoryBreakdown{
			{Category: CategoryImages, FileCount: 2, Bytes: 3000, Proportion: 1},
		},
	}

	if err := cache.Write(root, stats); err != nil {
		t.Fatal(err)
	}

	read, err := cache.Read(root)
	if err != nil {
		t.Fatal(err)
	}
	if read.FileCount != 2 || read.TotalBytes != 3000 {
		t.Fatalf("unexpected read: %+v", read)
	}

	path := filepath.Join(root, volume.CacheDir, volume.MetadataDir, volume.StatsCacheFile)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stats file missing: %v", err)
	}
}

func TestStatsCache_Invalidate(t *testing.T) {
	root := t.TempDir()
	cache := NewStatsCache()

	stats := cachedVolumeStats{ComputedAt: time.Now().UTC(), FileCount: 1}
	if err := cache.Write(root, stats); err != nil {
		t.Fatal(err)
	}
	if err := cache.Invalidate(root); err != nil {
		t.Fatal(err)
	}
	if _, err := cache.Read(root); !os.IsNotExist(err) {
		t.Fatalf("expected not exist after invalidate, got %v", err)
	}
}
