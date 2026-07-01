package volume

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func volumeJSONPath(rootPath string) string {
	return filepath.Join(rootPath, VolumeJSONFile)
}

func ReadVolumeConfig(rootPath string) (VolumeConfig, error) {
	data, err := os.ReadFile(volumeJSONPath(rootPath))
	if err != nil {
		return VolumeConfig{}, fmt.Errorf("read volume config: %w", err)
	}

	var cfg VolumeConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return VolumeConfig{}, fmt.Errorf("parse volume config: %w", err)
	}
	return cfg, nil
}

func WriteVolumeConfig(rootPath string, cfg VolumeConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal volume config: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(volumeJSONPath(rootPath))
	tmp, err := os.CreateTemp(dir, ".volume.json.*.tmp")
	if err != nil {
		return fmt.Errorf("create temp volume config: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("write temp volume config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("close temp volume config: %w", err)
	}

	if err := os.Rename(tmpName, volumeJSONPath(rootPath)); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("rename volume config: %w", err)
	}
	return nil
}

func SyncVolumeFromDisk(rootPath string) (*Volume, error) {
	cfg, err := ReadVolumeConfig(rootPath)
	if err != nil {
		return nil, err
	}
	return volumeFromConfig(cfg, rootPath), nil
}
