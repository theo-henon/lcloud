//go:build linux

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

// statBlockUnit returns the size of one statfs block count unit.
// Linux statfs(2): f_blocks and f_bavail are in f_frsize units, not f_bsize.
// Using f_bsize inflates totals on bind mounts (e.g. Docker Desktop virtiofs).
func statBlockUnit(stat *syscall.Statfs_t) uint64 {
	if stat.Frsize > 0 {
		return uint64(stat.Frsize)
	}
	if stat.Bsize > 0 {
		return uint64(stat.Bsize)
	}
	return 4096
}
