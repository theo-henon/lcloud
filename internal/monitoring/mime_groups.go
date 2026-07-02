package monitoring

import (
	"math"
	"strings"
)

// UsagePercentInt returns used/quota as a whole percent (0-100), rounded
// to match UI display (e.g. UnifiedVolumeList, VolumeStatsPanel).
func UsagePercentInt(usedBytes, quotaBytes int64) int {
	if quotaBytes <= 0 {
		return 0
	}
	return int(math.Round(float64(usedBytes) / float64(quotaBytes) * 100))
}

const (
	CategoryImages    = "images"
	CategoryVideos    = "videos"
	CategoryAudio     = "audio"
	CategoryDocuments = "documents"
	CategoryArchives  = "archives"
	CategoryOther     = "other"
)

func CategoryForMIME(mime string) string {
	mime = strings.ToLower(strings.TrimSpace(mime))
	if mime == "" {
		return CategoryOther
	}

	switch {
	case strings.HasPrefix(mime, "image/"):
		return CategoryImages
	case strings.HasPrefix(mime, "video/"):
		return CategoryVideos
	case strings.HasPrefix(mime, "audio/"):
		return CategoryAudio
	case strings.HasPrefix(mime, "text/"):
		return CategoryDocuments
	case mime == "application/pdf",
		mime == "application/msword",
		mime == "application/rtf":
		return CategoryDocuments
	case strings.HasPrefix(mime, "application/vnd."):
		return CategoryDocuments
	case mime == "application/zip",
		mime == "application/x-tar",
		mime == "application/gzip",
		mime == "application/x-7z-compressed",
		mime == "application/x-rar-compressed",
		strings.HasSuffix(mime, "+zip"):
		return CategoryArchives
	default:
		return CategoryOther
	}
}
