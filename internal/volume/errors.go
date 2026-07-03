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
	ErrUploadTooLarge     = errors.New("upload exceeds max size")
	ErrUploadSizeMismatch = errors.New("upload size mismatch")
	ErrPathTraversal      = errors.New("path traversal detected")
	ErrFileNotFound       = errors.New("file not found")
	ErrDirectoryExists    = errors.New("directory already exists")
	ErrDirectoryNotEmpty  = errors.New("directory not empty")
	ErrNotDirectory       = errors.New("not a directory")
	ErrNotAFile           = errors.New("not a file")
	ErrInvalidEntryName   = errors.New("invalid entry name")
	ErrForbidden                 = errors.New("forbidden")
	ErrDeletionRequestExists     = errors.New("deletion request already pending")
	ErrDeletionRequestNotFound   = errors.New("deletion request not found")
)
