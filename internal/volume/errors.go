package volume

import "errors"

var (
	ErrDiskNotFound       = errors.New("disk not found")
	ErrDiskNotWritable    = errors.New("disk not writable")
	ErrVolumeNotFound     = errors.New("volume not found")
	ErrVolumeNameTaken    = errors.New("volume name already taken")
	ErrVolumeNotEmpty     = errors.New("volume is not empty")
	ErrInvalidFilter      = errors.New("invalid filter configuration")
	ErrFileFilterRejected = errors.New("file extension not allowed")
	ErrQuotaExceeded      = errors.New("quota exceeded")
	ErrPathTraversal      = errors.New("path traversal detected")
	ErrFileNotFound       = errors.New("file not found")
	ErrDirectoryExists    = errors.New("directory already exists")
	ErrNotDirectory       = errors.New("not a directory")
	ErrForbidden          = errors.New("forbidden")
)
