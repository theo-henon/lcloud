package monitoring

import (
	"context"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/volume"
)

type Service struct {
	volumes  *volume.Service
	disks    *volume.DiskRegistry
	metadata *volume.MetadataCache
	cache    *StatsCache
	settings SettingsReader
}

type SettingsReader interface {
	ShouldMaskFor(claims *auth.Claims) (bool, error)
}

func NewService(volumes *volume.Service, disks *volume.DiskRegistry, settings SettingsReader) *Service {
	return &Service{
		volumes:  volumes,
		disks:    disks,
		metadata: volume.NewMetadataCache(),
		cache:    NewStatsCache(),
		settings: settings,
	}
}

func (s *Service) StatsCache() *StatsCache {
	return s.cache
}

func (s *Service) GetOverview(ctx context.Context, claims *auth.Claims) (*OverviewResponse, error) {
	_ = ctx

	disks := s.disks.ListDisks()
	volumes, err := s.volumes.List(claims, nil)
	if err != nil {
		return nil, err
	}

	usedByDisk := make(map[string]int64)
	volumeCountByDisk := make(map[string]int)
	for _, vol := range volumes {
		usedByDisk[vol.DiskPath] += vol.UsedBytes
		volumeCountByDisk[vol.DiskPath]++
	}

	diskOverviews := make([]DiskOverview, 0, len(disks))
	for _, disk := range disks {
		diskOverviews = append(diskOverviews, DiskOverview{
			Path:               disk.Path,
			Name:               disk.Name,
			Label:              disk.Label,
			TotalBytes:         disk.TotalBytes,
			FreeBytes:          disk.FreeBytes,
			UsedByVolumesBytes: usedByDisk[disk.Path],
			VolumeCount:        volumeCountByDisk[disk.Path],
		})
	}

	var ownerEmails map[uuid.UUID]string
	if claims.Role == auth.RoleAdmin {
		ownerIDs := make([]uuid.UUID, len(volumes))
		for i := range volumes {
			ownerIDs[i] = volumes[i].OwnerID
		}
		ownerEmails, err = s.volumes.OwnerEmailsByIDs(ownerIDs)
		if err != nil {
			return nil, err
		}
	}

	summaries := make([]VolumeSummary, 0, len(volumes))
	for i := range volumes {
		summary, err := s.volumeSummary(&volumes[i])
		if err != nil {
			return nil, err
		}
		if ownerEmails != nil {
			summary.OwnerEmail = ownerEmails[volumes[i].OwnerID]
		}
		summaries = append(summaries, *summary)
	}

	maskDiskNames := false
	if s.settings != nil {
		maskDiskNames, err = s.settings.ShouldMaskFor(claims)
		if err != nil {
			return nil, err
		}
	}

	if maskDiskNames {
		for i := range diskOverviews {
			diskOverviews[i] = maskOverviewDisk(diskOverviews[i], i)
		}
		for i := range summaries {
			summaries[i].DiskPath = ""
		}
	}

	return &OverviewResponse{
		Disks:   diskOverviews,
		Volumes: summaries,
	}, nil
}

func (s *Service) volumeSummary(vol *volume.Volume) (*VolumeSummary, error) {
	summary := &VolumeSummary{
		ID:         vol.ID.String(),
		Name:       vol.Name,
		DiskPath:   vol.DiskPath,
		QuotaBytes: vol.QuotaBytes,
		UsedBytes:  vol.UsedBytes,
	}

	cached, err := s.cache.Read(vol.RootPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		computed, computeErr := s.ComputeStats(vol)
		if computeErr != nil {
			return summary, nil
		}
		cached = cachedVolumeStats{
			ComputedAt: computed.ComputedAt,
			FileCount:  computed.FileCount,
			ByCategory: computed.ByCategory,
		}
	}

	summary.FileCount = cached.FileCount
	summary.TopCategory = topCategory(cached.ByCategory)
	computedAt := cached.ComputedAt
	summary.StatsComputedAt = &computedAt
	return summary, nil
}

func (s *Service) GetVolumeStats(ctx context.Context, claims *auth.Claims, volumeID uuid.UUID) (*VolumeStats, error) {
	_ = ctx

	vol, err := s.volumes.Get(claims, volumeID)
	if err != nil {
		return nil, err
	}

	if cached, err := s.cache.Read(vol.RootPath); err == nil {
		return enrichVolumeStats(vol, cached), nil
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	return s.ComputeStats(vol)
}

func (s *Service) ComputeStats(vol *volume.Volume) (*VolumeStats, error) {
	records, err := s.metadata.ListAll(vol.RootPath)
	if err != nil {
		return nil, err
	}

	cached := aggregateRecords(records, time.Now().UTC())
	if err := s.cache.Write(vol.RootPath, cached); err != nil {
		return nil, err
	}
	return enrichVolumeStats(vol, cached), nil
}

func (s *Service) RefreshVolumeStats(ctx context.Context, claims *auth.Claims, volumeID uuid.UUID) (*VolumeStats, error) {
	_ = ctx

	vol, err := s.volumes.Get(claims, volumeID)
	if err != nil {
		return nil, err
	}
	return s.ComputeStats(vol)
}
