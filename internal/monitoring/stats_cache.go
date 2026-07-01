package monitoring

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/theo-henon/lcloud/internal/volume"
)

type StatsCache struct {
	metadata *volume.MetadataCache
}

func NewStatsCache() *StatsCache {
	return &StatsCache{metadata: volume.NewMetadataCache()}
}

func (c *StatsCache) statsPath(rootPath string) string {
	return filepath.Join(rootPath, volume.CacheDir, volume.MetadataDir, volume.StatsCacheFile)
}

func (c *StatsCache) Read(rootPath string) (cachedVolumeStats, error) {
	data, err := os.ReadFile(c.statsPath(rootPath))
	if err != nil {
		return cachedVolumeStats{}, err
	}
	var cached cachedVolumeStats
	if err := json.Unmarshal(data, &cached); err != nil {
		return cachedVolumeStats{}, err
	}
	return cached, nil
}

func (c *StatsCache) Write(rootPath string, stats cachedVolumeStats) error {
	dir := filepath.Join(rootPath, volume.CacheDir, volume.MetadataDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmpPath := c.statsPath(rootPath) + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpPath, c.statsPath(rootPath))
}

func (c *StatsCache) Invalidate(rootPath string) error {
	err := os.Remove(c.statsPath(rootPath))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
