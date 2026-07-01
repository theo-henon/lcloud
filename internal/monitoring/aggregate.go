package monitoring

import (
	"sort"
	"time"

	"github.com/theo-henon/lcloud/internal/volume"
)

func aggregateRecords(records []volume.FileMetadataRecord, computedAt time.Time) cachedVolumeStats {
	categoryMap := make(map[string]*CategoryBreakdown)
	mimeMap := make(map[string]*MimeBreakdown)
	var totalBytes int64

	for _, record := range records {
		totalBytes += record.SizeBytes
		category := CategoryForMIME(record.MimeType)

		if cat, ok := categoryMap[category]; ok {
			cat.FileCount++
			cat.Bytes += record.SizeBytes
		} else {
			categoryMap[category] = &CategoryBreakdown{
				Category:  category,
				FileCount: 1,
				Bytes:     record.SizeBytes,
			}
		}

		if mime, ok := mimeMap[record.MimeType]; ok {
			mime.FileCount++
			mime.Bytes += record.SizeBytes
		} else {
			mimeMap[record.MimeType] = &MimeBreakdown{
				MimeType:  record.MimeType,
				FileCount: 1,
				Bytes:     record.SizeBytes,
			}
		}
	}

	byCategory := make([]CategoryBreakdown, 0, len(categoryMap))
	for _, item := range categoryMap {
		if totalBytes > 0 {
			item.Proportion = float64(item.Bytes) / float64(totalBytes)
		}
		byCategory = append(byCategory, *item)
	}
	sort.Slice(byCategory, func(i, j int) bool {
		return byCategory[i].Bytes > byCategory[j].Bytes
	})

	byMime := make([]MimeBreakdown, 0, len(mimeMap))
	for _, item := range mimeMap {
		if totalBytes > 0 {
			item.Proportion = float64(item.Bytes) / float64(totalBytes)
		}
		byMime = append(byMime, *item)
	}
	sort.Slice(byMime, func(i, j int) bool {
		return byMime[i].Bytes > byMime[j].Bytes
	})

	return cachedVolumeStats{
		ComputedAt: computedAt,
		FileCount:  len(records),
		TotalBytes: totalBytes,
		ByCategory: byCategory,
		ByMime:     byMime,
	}
}

func topCategory(byCategory []CategoryBreakdown) string {
	if len(byCategory) == 0 {
		return ""
	}
	return byCategory[0].Category
}

func enrichVolumeStats(vol *volume.Volume, cached cachedVolumeStats) *VolumeStats {
	stats := &VolumeStats{
		VolumeID:   vol.ID.String(),
		Name:       vol.Name,
		QuotaBytes: vol.QuotaBytes,
		UsedBytes:  vol.UsedBytes,
		FileCount:  cached.FileCount,
		ComputedAt: cached.ComputedAt,
		ByCategory: cached.ByCategory,
		ByMime:     cached.ByMime,
	}

	if vol.QuotaBytes > 0 {
		free := vol.QuotaBytes - vol.UsedBytes
		if free < 0 {
			free = 0
		}
		stats.FreeBytes = &free
		pct := float64(vol.UsedBytes) / float64(vol.QuotaBytes)
		stats.UsagePercent = &pct
	}

	return stats
}
