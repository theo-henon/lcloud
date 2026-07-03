package ftp

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/google/uuid"
	"github.com/spf13/afero"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/protocols"
	"github.com/theo-henon/lcloud/internal/volume"
)

type volumeFs struct {
	gateway  *protocols.Gateway
	claims   *auth.Claims
	volumeID uuid.UUID
}

func newVolumeFs(gateway *protocols.Gateway, claims *auth.Claims, volumeID uuid.UUID) *volumeFs {
	return &volumeFs{gateway: gateway, claims: claims, volumeID: volumeID}
}

func (fs *volumeFs) clean(name string) (string, error) {
	name = path.Clean("/" + strings.ReplaceAll(name, "\\", "/"))
	name = strings.TrimPrefix(name, "/")
	if name == "" || name == "." {
		return ".", nil
	}
	if strings.Contains(name, "..") {
		return "", volume.ErrPathTraversal
	}
	return name, nil
}

func (fs *volumeFs) Create(name string) (afero.File, error) {
	return fs.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
}

func (fs *volumeFs) Mkdir(name string, perm os.FileMode) error {
	rel, err := fs.clean(name)
	if err != nil {
		return err
	}
	return fs.gateway.Files().CreateDirectory(fs.claims, fs.volumeID, rel)
}

func (fs *volumeFs) MkdirAll(path string, perm os.FileMode) error {
	return fs.Mkdir(path, perm)
}

func (fs *volumeFs) Open(name string) (afero.File, error) {
	return fs.OpenFile(name, os.O_RDONLY, 0)
}

func (fs *volumeFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if flag&(os.O_WRONLY|os.O_RDWR|os.O_CREATE|os.O_APPEND) != 0 {
		return nil, os.ErrPermission
	}
	rel, err := fs.clean(name)
	if err != nil {
		return nil, err
	}
	file, _, err := fs.gateway.Files().OpenContent(fs.claims, fs.volumeID, rel)
	if err != nil {
		return nil, mapFTPError(err)
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	return &ftpRead{File: file, info: info}, nil
}

func (fs *volumeFs) GetHandle(name string, flags int, offset int64) (ftpserver.FileTransfer, error) {
	rel, err := fs.clean(name)
	if err != nil {
		return nil, err
	}

	if flags&os.O_RDONLY != 0 {
		file, _, err := fs.gateway.Files().OpenContent(fs.claims, fs.volumeID, rel)
		if err != nil {
			return nil, mapFTPError(err)
		}
		if offset > 0 {
			if _, err := file.Seek(offset, io.SeekStart); err != nil {
				_ = file.Close()
				return nil, err
			}
		}
		return file, nil
	}

	dir := filepath.Dir(rel)
	if dir == "." {
		dir = "."
	}
	return &ftpTransfer{
		fs:       fs,
		dir:      dir,
		filename: filepath.Base(rel),
	}, nil
}

func (fs *volumeFs) Remove(name string) error {
	rel, err := fs.clean(name)
	if err != nil {
		return err
	}
	return mapFTPError(fs.gateway.Files().Delete(fs.claims, fs.volumeID, rel))
}

func (fs *volumeFs) RemoveDir(name string) error {
	rel, err := fs.clean(name)
	if err != nil {
		return err
	}
	return mapFTPError(fs.gateway.Files().RemoveDirectory(fs.claims, fs.volumeID, rel))
}

func (fs *volumeFs) RemoveAll(name string) error {
	rel, err := fs.clean(name)
	if err != nil {
		return err
	}
	if rel == "." {
		return os.ErrPermission
	}
	listing, err := fs.gateway.Files().List(fs.claims, fs.volumeID, rel)
	if err != nil {
		if errors.Is(err, volume.ErrNotDirectory) {
			return mapFTPError(fs.gateway.Files().Delete(fs.claims, fs.volumeID, rel))
		}
		return mapFTPError(err)
	}
	for _, entry := range listing.Entries {
		if err := fs.RemoveAll("/" + entry.Path); err != nil {
			return err
		}
	}
	return mapFTPError(fs.gateway.Files().RemoveDirectory(fs.claims, fs.volumeID, rel))
}

func (fs *volumeFs) Rename(oldname, newname string) error {
	oldRel, err := fs.clean(oldname)
	if err != nil {
		return err
	}
	newRel, err := fs.clean(newname)
	if err != nil {
		return err
	}
	if filepath.Dir(oldRel) == filepath.Dir(newRel) {
		_, err := fs.gateway.Files().Rename(fs.claims, fs.volumeID, oldRel, filepath.Base(newRel))
		return mapFTPError(err)
	}
	_, err = fs.gateway.Files().Move(fs.claims, fs.volumeID, oldRel, newRel)
	return mapFTPError(err)
}

func (fs *volumeFs) Stat(name string) (os.FileInfo, error) {
	rel, err := fs.clean(name)
	if err != nil {
		return nil, err
	}
	if rel == "." {
		return ftpDirInfo{name: "/"}, nil
	}
	parent := filepath.Dir(rel)
	if parent == "." {
		parent = "."
	}
	listing, err := fs.gateway.Files().List(fs.claims, fs.volumeID, parent)
	if err != nil {
		return nil, mapFTPError(err)
	}
	base := filepath.Base(rel)
	for _, entry := range listing.Entries {
		if entry.Name == base {
			return ftpEntryInfo{entry: entry}, nil
		}
	}
	return nil, os.ErrNotExist
}

func (fs *volumeFs) Name() string { return "volume" }
func (fs *volumeFs) Chmod(string, os.FileMode) error { return os.ErrPermission }
func (fs *volumeFs) Chown(string, int, int) error    { return os.ErrPermission }
func (fs *volumeFs) Chtimes(string, time.Time, time.Time) error { return os.ErrPermission }

func (fs *volumeFs) ReadDir(name string) ([]os.FileInfo, error) {
	rel, err := fs.clean(name)
	if err != nil {
		return nil, err
	}
	listing, err := fs.gateway.Files().List(fs.claims, fs.volumeID, rel)
	if err != nil {
		return nil, mapFTPError(err)
	}
	out := make([]os.FileInfo, 0, len(listing.Entries))
	for _, entry := range listing.Entries {
		out = append(out, ftpEntryInfo{entry: entry})
	}
	return out, nil
}

type ftpRead struct {
	*os.File
	info os.FileInfo
}

func (f *ftpRead) Readdir(count int) ([]os.FileInfo, error) {
	return nil, os.ErrInvalid
}

type ftpTransfer struct {
	fs       *volumeFs
	dir      string
	filename string
	data     []byte
	closed   bool
}

func (f *ftpTransfer) Read(p []byte) (int, error) {
	return 0, io.EOF
}

func (f *ftpTransfer) Write(p []byte) (int, error) {
	if f.closed {
		return 0, os.ErrClosed
	}
	f.data = append(f.data, p...)
	return len(p), nil
}

func (f *ftpTransfer) Seek(offset int64, whence int) (int64, error) {
	return 0, os.ErrInvalid
}

func (f *ftpTransfer) Close() error {
	if f.closed {
		return nil
	}
	f.closed = true
	if len(f.data) == 0 {
		return nil
	}
	_, err := f.fs.gateway.Files().Upload(
		f.fs.claims,
		f.fs.volumeID,
		f.dir,
		f.filename,
		bytes.NewReader(f.data),
		int64(len(f.data)),
	)
	return mapFTPError(err)
}

type ftpEntryInfo struct {
	entry volume.FileEntry
}

func (e ftpEntryInfo) Name() string { return e.entry.Name }
func (e ftpEntryInfo) Size() int64 {
	if e.entry.Type == "directory" {
		return 0
	}
	return e.entry.SizeBytes
}
func (e ftpEntryInfo) Mode() os.FileMode {
	if e.entry.Type == "directory" {
		return os.ModeDir | 0o755
	}
	return 0o644
}
func (e ftpEntryInfo) ModTime() time.Time {
	if e.entry.ModifiedAt.IsZero() {
		return time.Time{}
	}
	return e.entry.ModifiedAt
}
func (e ftpEntryInfo) IsDir() bool { return e.entry.Type == "directory" }
func (e ftpEntryInfo) Sys() any    { return nil }

type ftpDirInfo struct {
	name string
}

func (d ftpDirInfo) Name() string       { return d.name }
func (d ftpDirInfo) Size() int64        { return 0 }
func (d ftpDirInfo) Mode() os.FileMode  { return os.ModeDir | 0o755 }
func (d ftpDirInfo) ModTime() time.Time { return time.Time{} }
func (d ftpDirInfo) IsDir() bool        { return true }
func (d ftpDirInfo) Sys() any           { return nil }

func mapFTPError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, volume.ErrFileNotFound) || errors.Is(err, volume.ErrVolumeNotFound) {
		return os.ErrNotExist
	}
	return err
}

var _ ftpserver.ClientDriver = (*volumeFs)(nil)
var _ ftpserver.ClientDriverExtensionFileList = (*volumeFs)(nil)
var _ ftpserver.ClientDriverExtensionRemoveDir = (*volumeFs)(nil)
var _ ftpserver.ClientDriverExtentionFileTransfer = (*volumeFs)(nil)
