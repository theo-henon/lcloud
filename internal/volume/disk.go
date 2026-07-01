package volume

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/theo-henon/lcloud/internal/config"
)

type DiskInfo struct {
	Path       string `json:"path"`
	Name       string `json:"name"`
	Label      string `json:"label"`
	TotalBytes uint64 `json:"total_bytes"`
	FreeBytes  uint64 `json:"free_bytes"`
}

type DiskRegistry struct {
	cfg *config.Config
}

func NewDiskRegistry(cfg *config.Config) *DiskRegistry {
	return &DiskRegistry{cfg: cfg}
}

func (r *DiskRegistry) ResolvedPaths() []string {
	if len(r.cfg.StorageDiskPaths) > 0 {
		return r.cfg.StorageDiskPaths
	}
	return r.discoverFromBasePath()
}

func (r *DiskRegistry) discoverFromBasePath() []string {
	base := r.cfg.StorageBasePath
	entries, err := os.ReadDir(base)
	if err != nil {
		log.Printf("disk registry: unable to scan STORAGE_BASE_PATH %q: %v", base, err)
		return nil
	}

	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		paths = append(paths, filepath.Join(base, entry.Name()))
	}
	if len(paths) == 0 {
		if info, err := os.Stat(base); err == nil && info.IsDir() {
			return []string{base}
		}
	}
	return paths
}

func (r *DiskRegistry) ListDisks() []DiskInfo {
	paths := r.ResolvedPaths()
	disks := make([]DiskInfo, 0, len(paths))

	for i, path := range paths {
		info, ok := r.diskInfo(path, i)
		if !ok {
			log.Printf("disk registry: skipping invalid disk path %q", path)
			continue
		}
		disks = append(disks, info)
	}
	return disks
}

func (r *DiskRegistry) diskInfo(path string, index int) (DiskInfo, bool) {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return DiskInfo{}, false
	}

	if err := checkWritable(path); err != nil {
		return DiskInfo{}, false
	}

	total, free := diskUsage(path)
	name := filepath.Base(path)
	label := name
	if index < len(r.cfg.StorageDiskLabels) {
		label = r.cfg.StorageDiskLabels[index]
	}

	return DiskInfo{
		Path:       path,
		Name:       name,
		Label:      label,
		TotalBytes: total,
		FreeBytes:  free,
	}, true
}

func (r *DiskRegistry) IsRegistered(path string) bool {
	clean := filepath.Clean(path)
	for _, diskPath := range r.ResolvedPaths() {
		if filepath.Clean(diskPath) == clean {
			info, err := os.Stat(diskPath)
			return err == nil && info.IsDir() && checkWritable(diskPath) == nil
		}
	}
	return false
}

func checkWritable(path string) error {
	testFile := filepath.Join(path, ".lcloud-write-test")
	if err := os.WriteFile(testFile, []byte("ok"), 0o644); err != nil {
		return err
	}
	return os.Remove(testFile)
}

func diskNameFromPath(path string) string {
	return strings.TrimSpace(filepath.Base(filepath.Clean(path)))
}
