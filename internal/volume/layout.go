package volume

import (
	"os"
	"path/filepath"
)

func createVolumeLayout(rootPath string) error {
	dirs := []string{
		filepath.Join(rootPath, UserdataDir),
		filepath.Join(rootPath, CacheDir),
		filepath.Join(rootPath, CacheDir, ThumbnailsDir),
		filepath.Join(rootPath, CacheDir, MetadataDir),
		filepath.Join(rootPath, PluginsDir),
		filepath.Join(rootPath, LogsDir),
		filepath.Join(rootPath, TrashDir),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func isUserdataEmpty(rootPath string) (bool, error) {
	userdataRoot := NewPathResolver().UserdataRoot(rootPath)
	entries, err := os.ReadDir(userdataRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	return len(entries) == 0, nil
}

func removeVolumeTree(rootPath string) error {
	return os.RemoveAll(rootPath)
}
