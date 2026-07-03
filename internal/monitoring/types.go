package monitoring

import "time"

type CategoryBreakdown struct {
	Category   string  `json:"category"`
	FileCount  int     `json:"file_count"`
	Bytes      int64   `json:"bytes"`
	Proportion float64 `json:"proportion"`
}

type MimeBreakdown struct {
	MimeType   string  `json:"mime_type"`
	FileCount  int     `json:"file_count"`
	Bytes      int64   `json:"bytes"`
	Proportion float64 `json:"proportion"`
}

type VolumeStats struct {
	VolumeID      string              `json:"volume_id"`
	Name          string              `json:"name"`
	QuotaBytes    int64               `json:"quota_bytes"`
	UsedBytes     int64               `json:"used_bytes"`
	FreeBytes     *int64              `json:"free_bytes,omitempty"`
	UsagePercent  *float64            `json:"usage_percent,omitempty"`
	FileCount     int                 `json:"file_count"`
	ComputedAt    time.Time           `json:"computed_at"`
	ByCategory    []CategoryBreakdown `json:"by_category"`
	ByMime        []MimeBreakdown     `json:"by_mime"`
}

type cachedVolumeStats struct {
	ComputedAt time.Time           `json:"computed_at"`
	FileCount  int                 `json:"file_count"`
	TotalBytes int64               `json:"total_bytes"`
	ByCategory []CategoryBreakdown `json:"by_category"`
	ByMime     []MimeBreakdown     `json:"by_mime"`
}

type VolumeSummary struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	OwnerEmail      string     `json:"owner_email,omitempty"`
	DiskPath        string     `json:"disk_path"`
	QuotaBytes      int64      `json:"quota_bytes"`
	UsedBytes       int64      `json:"used_bytes"`
	FileCount       int         `json:"file_count"`
	TopCategory     string     `json:"top_category,omitempty"`
	StatsComputedAt *time.Time `json:"stats_computed_at,omitempty"`
}

type DiskOverview struct {
	Path                 string `json:"path"`
	Name                 string `json:"name"`
	Label                string `json:"label"`
	TotalBytes           uint64 `json:"total_bytes"`
	FreeBytes            uint64 `json:"free_bytes"`
	UsedByVolumesBytes   int64  `json:"used_by_volumes_bytes"`
	VolumeCount          int    `json:"volume_count"`
}

type OverviewResponse struct {
	Disks   []DiskOverview  `json:"disks"`
	Volumes []VolumeSummary `json:"volumes"`
}
