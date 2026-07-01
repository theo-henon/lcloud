//go:build !windows && !linux

package volume

import "syscall"

func diskUsage(path string) (total, free uint64) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, 0
	}
	unit := statBlockUnit(&stat)
	total = stat.Blocks * unit
	free = stat.Bavail * unit
	return total, free
}

func statBlockUnit(stat *syscall.Statfs_t) uint64 {
	if stat.Bsize > 0 {
		return uint64(stat.Bsize)
	}
	return 4096
}
