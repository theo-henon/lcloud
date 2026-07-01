//go:build windows

package volume

func diskUsage(path string) (total, free uint64) {
	return 0, 0
}
