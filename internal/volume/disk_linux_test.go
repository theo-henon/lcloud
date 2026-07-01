//go:build linux

package volume

import (
	"syscall"
	"testing"
)

func TestStatBlockUnit_PrefersFrsize(t *testing.T) {
	stat := syscall.Statfs_t{
		Bsize:  1048576,
		Frsize: 4096,
	}
	if got := statBlockUnit(&stat); got != 4096 {
		t.Fatalf("statBlockUnit() = %d, want 4096", got)
	}
}

func TestStatBlockUnit_FallsBackToBsize(t *testing.T) {
	stat := syscall.Statfs_t{
		Bsize: 4096,
	}
	if got := statBlockUnit(&stat); got != 4096 {
		t.Fatalf("statBlockUnit() = %d, want 4096", got)
	}
}

func TestDiskUsage_SaneRatio(t *testing.T) {
	total, free := diskUsage(t.TempDir())
	if total == 0 {
		t.Skip("statfs unavailable in this environment")
	}
	if free > total {
		t.Fatalf("free (%d) exceeds total (%d)", free, total)
	}
}
