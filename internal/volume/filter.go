package volume

import (
	"path/filepath"
	"strings"
)

func NormalizeExtension(ext string) string {
	ext = strings.TrimSpace(strings.ToLower(ext))
	if ext == "" {
		return ""
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return ext
}

func NormalizeExtensions(extensions []string) []string {
	out := make([]string, 0, len(extensions))
	seen := make(map[string]struct{}, len(extensions))
	for _, ext := range extensions {
		normalized := NormalizeExtension(ext)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func NormalizeFilters(filters Filters) Filters {
	filters.Extensions = NormalizeExtensions(filters.Extensions)
	if len(filters.Extensions) == 0 {
		filters.Mode = ""
	}
	return filters
}

func ValidateFilters(filters Filters) error {
	switch filters.Mode {
	case "", FilterModeAllow, FilterModeBlock:
		return nil
	default:
		return ErrInvalidFilter
	}
}

func ValidateExtension(filters Filters, filename string) error {
	if filters.Mode == "" || len(filters.Extensions) == 0 {
		return nil
	}

	ext := NormalizeExtension(filepath.Ext(filename))
	allowedSet := make(map[string]struct{}, len(filters.Extensions))
	for _, item := range NormalizeExtensions(filters.Extensions) {
		allowedSet[item] = struct{}{}
	}

	_, matches := allowedSet[ext]
	switch filters.Mode {
	case FilterModeAllow:
		if !matches {
			return ErrFileFilterRejected
		}
	case FilterModeBlock:
		if matches {
			return ErrFileFilterRejected
		}
	}
	return nil
}
