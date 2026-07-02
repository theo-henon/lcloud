package httputil

import (
	"fmt"
	"path/filepath"
	"strings"
)

func ContentDispositionAttachment(filename string) string {
	safe := sanitizeFilename(filepath.Base(filename))
	if safe == "" {
		safe = "download"
	}
	return fmt.Sprintf("attachment; filename=%q", safe)
}

func ContentDispositionInline(filename string) string {
	safe := sanitizeFilename(filepath.Base(filename))
	if safe == "" {
		safe = "preview"
	}
	return fmt.Sprintf("inline; filename=%q", safe)
}

func sanitizeFilename(name string) string {
	var b strings.Builder
	b.Grow(len(name))
	for _, r := range name {
		if r <= 31 || r == '"' || r == '\\' {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
