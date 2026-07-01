package monitoring

import "fmt"

func maskOverviewDisk(disk DiskOverview, index int) DiskOverview {
	label := fmt.Sprintf("Storage %d", index+1)
	return DiskOverview{
		Path:                 "",
		Name:                 fmt.Sprintf("storage-%d", index+1),
		Label:                label,
		TotalBytes:           disk.TotalBytes,
		FreeBytes:            disk.FreeBytes,
		UsedByVolumesBytes:   disk.UsedByVolumesBytes,
		VolumeCount:          disk.VolumeCount,
	}
}
