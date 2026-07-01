package api

import (
	"fmt"

	"github.com/theo-henon/lcloud/internal/volume"
)

func maskDiskInfo(disk volume.DiskInfo, index int) volume.DiskInfo {
	label := fmt.Sprintf("Storage %d", index+1)
	return volume.DiskInfo{
		Path:       "",
		Name:       fmt.Sprintf("storage-%d", index+1),
		Label:      label,
		TotalBytes: disk.TotalBytes,
		FreeBytes:  disk.FreeBytes,
	}
}

func maskVolume(vol volume.Volume) volume.Volume {
	vol.DiskPath = ""
	return vol
}

func maskVolumePtr(vol *volume.Volume) *volume.Volume {
	if vol == nil {
		return nil
	}
	masked := maskVolume(*vol)
	return &masked
}
