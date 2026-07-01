package volume

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type PathResolver struct{}

func NewPathResolver() *PathResolver {
	return &PathResolver{}
}

func (p *PathResolver) UserdataRoot(rootPath string) string {
	return filepath.Join(rootPath, UserdataDir)
}

func (p *PathResolver) ResolveUserdata(rootPath, relPath string) (string, error) {
	if strings.Contains(relPath, "..") {
		return "", ErrPathTraversal
	}
	if strings.HasPrefix(strings.TrimSpace(relPath), "/") {
		return "", ErrPathTraversal
	}

	clean := cleanRelativePath(relPath)
	if clean == "" {
		clean = "."
	}

	userdataRoot, err := filepath.Abs(p.UserdataRoot(rootPath))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrPathTraversal, err)
	}

	target := filepath.Join(userdataRoot, clean)
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrPathTraversal, err)
	}

	if !pathWithinRoot(userdataRoot, absTarget) {
		return "", ErrPathTraversal
	}

	if err := ensureNoEscapingSymlink(userdataRoot, absTarget); err != nil {
		return "", err
	}

	return absTarget, nil
}

func cleanRelativePath(relPath string) string {
	relPath = strings.TrimSpace(relPath)
	relPath = strings.TrimPrefix(relPath, "/")
	relPath = filepath.Clean(relPath)
	if relPath == "." {
		return "."
	}
	return relPath
}

func pathWithinRoot(root, target string) bool {
	root = filepath.Clean(root)
	target = filepath.Clean(target)
	if root == target {
		return true
	}
	sep := string(os.PathSeparator)
	return strings.HasPrefix(target, root+sep)
}

func ensureNoEscapingSymlink(root, target string) error {
	current := root
	parts := strings.Split(strings.TrimPrefix(filepath.Clean(target), filepath.Clean(root)), string(os.PathSeparator))
	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			resolved, err := filepath.EvalSymlinks(current)
			if err != nil {
				return fmt.Errorf("%w: %v", ErrPathTraversal, err)
			}
			if !pathWithinRoot(root, resolved) {
				return ErrPathTraversal
			}
			current = resolved
		}
	}
	return nil
}

func relativeUserdataPath(rootPath, absPath string) (string, error) {
	userdataRoot, err := filepath.Abs(NewPathResolver().UserdataRoot(rootPath))
	if err != nil {
		return "", err
	}
	absPath, err = filepath.Abs(absPath)
	if err != nil {
		return "", err
	}
	if !pathWithinRoot(userdataRoot, absPath) {
		return "", ErrPathTraversal
	}
	rel, err := filepath.Rel(userdataRoot, absPath)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(rel), nil
}
