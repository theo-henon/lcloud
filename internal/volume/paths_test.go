package volume

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveUserdataRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	resolver := NewPathResolver()
	require.NoError(t, os.MkdirAll(resolver.UserdataRoot(root), 0o755))

	_, err := resolver.ResolveUserdata(root, "../secret.txt")
	require.ErrorIs(t, err, ErrPathTraversal)

	_, err = resolver.ResolveUserdata(root, "/etc/passwd")
	require.ErrorIs(t, err, ErrPathTraversal)
}

func TestResolveUserdataAllowsNestedPath(t *testing.T) {
	root := t.TempDir()
	resolver := NewPathResolver()
	userdata := resolver.UserdataRoot(root)
	require.NoError(t, os.MkdirAll(filepath.Join(userdata, "2024"), 0o755))

	abs, err := resolver.ResolveUserdata(root, "2024/vacation.jpg")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(userdata, "2024", "vacation.jpg"), abs)
}

func TestResolveUserdataRejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	resolver := NewPathResolver()
	userdata := resolver.UserdataRoot(root)
	require.NoError(t, os.MkdirAll(userdata, 0o755))

	outside := t.TempDir()
	outsideFile := filepath.Join(outside, "secret.txt")
	require.NoError(t, os.WriteFile(outsideFile, []byte("secret"), 0o644))
	require.NoError(t, os.Symlink(outsideFile, filepath.Join(userdata, "link.txt")))

	_, err := resolver.ResolveUserdata(root, "link.txt")
	require.ErrorIs(t, err, ErrPathTraversal)
}

func TestRelativeUserdataPath(t *testing.T) {
	root := t.TempDir()
	resolver := NewPathResolver()
	userdata := resolver.UserdataRoot(root)
	require.NoError(t, os.MkdirAll(userdata, 0o755))

	abs := filepath.Join(userdata, "photos", "a.jpg")
	rel, err := relativeUserdataPath(root, abs)
	require.NoError(t, err)
	require.Equal(t, "photos/a.jpg", rel)
}
