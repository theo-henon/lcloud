package api

import (
	"fmt"

	"github.com/theo-henon/lcloud/internal/volume"
)

func maskDiskInfo(disk volume.DiskInfo, index int) volume.DiskInfo {
	label := fmt.Sprintf("Storage %d", index+1)
	id := volume.DiskIDForIndex(index)
	if disk.ID != "" {
		id = disk.ID
	}
	return volume.DiskInfo{
		ID:         id,
		Path:       "",
		Name:       id,
		Label:      label,
		TotalBytes: disk.TotalBytes,
		FreeBytes:  disk.FreeBytes,
	}
}

func maskVolume(vol volume.Volume) volume.Volume {
	vol.DiskPath = ""
	vol.RootPath = ""
	return vol
}
