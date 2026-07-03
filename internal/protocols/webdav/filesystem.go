package webdav

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/volume"
	"golang.org/x/net/webdav"
)

type volumeFS struct {
	files    *volume.FileService
	claims   *auth.Claims
	volumeID uuid.UUID
}

func newVolumeFS(files *volume.FileService, claims *auth.Claims, volumeID uuid.UUID) webdav.FileSystem {
	return &volumeFS{files: files, claims: claims, volumeID: volumeID}
}

func (fs *volumeFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	rel, err := fs.clean(name)
	if err != nil {
		return err
	}
	return fs.files.CreateDirectory(fs.claims, fs.volumeID, rel)
}

func (fs *volumeFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	rel, err := fs.clean(name)
	if err != nil {
		return nil, err
	}

	if flag&os.O_CREATE != 0 {
		dir := filepath.Dir(rel)
		if dir == "." {
			dir = "."
		}
		filename := filepath.Base(rel)
		return &uploadFile{
			fs:       fs,
			dir:      dir,
			filename: filename,
		}, nil
	}

	file, mimeType, err := fs.files.OpenContent(fs.claims, fs.volumeID, rel)
	if err != nil {
		return nil, mapError(err)
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	return &readFile{File: file, info: info, mimeType: mimeType}, nil
}

func (fs *volumeFS) RemoveAll(ctx context.Context, name string) error {
	rel, err := fs.clean(name)
	if err != nil {
		return err
	}
	if rel == "." {
		return os.ErrPermission
	}

	listing, err := fs.files.List(fs.claims, fs.volumeID, rel)
	if err != nil {
		if errors.Is(err, volume.ErrNotDirectory) {
			return fs.files.Delete(fs.claims, fs.volumeID, rel)
		}
		return mapError(err)
	}

	for _, entry := range listing.Entries {
		child := entry.Path
		if err := fs.RemoveAll(ctx, "/"+child); err != nil {
			return err
		}
	}
	return fs.files.RemoveDirectory(fs.claims, fs.volumeID, rel)
}

func (fs *volumeFS) Rename(ctx context.Context, oldName, newName string) error {
	oldRel, err := fs.clean(oldName)
	if err != nil {
		return err
	}
	newRel, err := fs.clean(newName)
	if err != nil {
		return err
	}

	oldDir := filepath.Dir(oldRel)
	newDir := filepath.Dir(newRel)
	if oldDir == newDir {
		newBase := filepath.Base(newRel)
		_, err := fs.files.Rename(fs.claims, fs.volumeID, oldRel, newBase)
		return mapError(err)
	}

	_, err = fs.files.Move(fs.claims, fs.volumeID, oldRel, newRel)
	return mapError(err)
}

func (fs *volumeFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	rel, err := fs.clean(name)
	if err != nil {
		return nil, err
	}
	if rel == "." {
		return rootDirInfo{}, nil
	}

	parent := filepath.Dir(rel)
	if parent == "." {
		parent = "."
	}
	base := filepath.Base(rel)

	listing, err := fs.files.List(fs.claims, fs.volumeID, parent)
	if err != nil {
		return nil, mapError(err)
	}
	for _, entry := range listing.Entries {
		if entry.Name == base {
			return entryInfo{entry: entry}, nil
		}
	}
	return nil, os.ErrNotExist
}

func (fs *volumeFS) clean(name string) (string, error) {
	name = path.Clean("/" + name)
	name = strings.TrimPrefix(name, "/")
	if name == "" || name == "." {
		return ".", nil
	}
	if strings.Contains(name, "..") {
		return "", volume.ErrPathTraversal
	}
	return name, nil
}

type rootDirInfo struct{}

func (rootDirInfo) Name() string       { return "/" }
func (rootDirInfo) Size() int64        { return 0 }
func (rootDirInfo) Mode() os.FileMode  { return os.ModeDir | 0o755 }
func (rootDirInfo) ModTime() time.Time { return time.Time{} }
func (rootDirInfo) IsDir() bool        { return true }
func (rootDirInfo) Sys() any           { return nil }

type entryInfo struct {
	entry volume.FileEntry
}

func (e entryInfo) Name() string { return e.entry.Name }
func (e entryInfo) Size() int64 {
	if e.entry.Type == "directory" {
		return 0
	}
	return e.entry.SizeBytes
}
func (e entryInfo) Mode() os.FileMode {
	if e.entry.Type == "directory" {
		return os.ModeDir | 0o755
	}
	return 0o644
}
func (e entryInfo) ModTime() time.Time {
	if e.entry.ModifiedAt.IsZero() {
		return time.Time{}
	}
	return e.entry.ModifiedAt
}
func (e entryInfo) IsDir() bool { return e.entry.Type == "directory" }
func (e entryInfo) Sys() any    { return nil }

type readFile struct {
	*os.File
	info     os.FileInfo
	mimeType string
}

func (f *readFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, os.ErrInvalid
}

type uploadFile struct {
	fs       *volumeFS
	dir      string
	filename string
	buffer   []byte
	closed   bool
}

func (f *uploadFile) Read(p []byte) (int, error) {
	return 0, io.EOF
}

func (f *uploadFile) Write(p []byte) (int, error) {
	if f.closed {
		return 0, os.ErrClosed
	}
	f.buffer = append(f.buffer, p...)
	return len(p), nil
}

func (f *uploadFile) Close() error {
	if f.closed {
		return nil
	}
	f.closed = true
	if len(f.buffer) == 0 {
		return nil
	}
	_, err := f.fs.files.Upload(f.fs.claims, f.fs.volumeID, f.dir, f.filename, bytes.NewReader(f.buffer), int64(len(f.buffer)))
	return mapError(err)
}

func (f *uploadFile) Seek(offset int64, whence int) (int64, error) {
	return 0, os.ErrInvalid
}

func (f *uploadFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, os.ErrInvalid
}

func (f *uploadFile) Stat() (os.FileInfo, error) {
	return entryInfo{entry: volume.FileEntry{Name: f.filename, Type: "file", SizeBytes: int64(len(f.buffer))}}, nil
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, volume.ErrPathTraversal):
		return os.ErrPermission
	case errors.Is(err, volume.ErrQuotaExceeded):
		return errQuotaExceeded
	case errors.Is(err, volume.ErrFileFilterRejected):
		return errUnsupportedMedia
	case errors.Is(err, volume.ErrForbidden):
		return os.ErrPermission
	case errors.Is(err, volume.ErrVolumeNotFound), errors.Is(err, volume.ErrFileNotFound):
		return os.ErrNotExist
	case errors.Is(err, volume.ErrDirectoryExists):
		return os.ErrExist
	case errors.Is(err, volume.ErrDirectoryNotEmpty):
		return os.ErrPermission
	default:
		return err
	}
}

var (
	errQuotaExceeded    = errors.New("quota exceeded")
	errUnsupportedMedia = errors.New("unsupported media type")
)
