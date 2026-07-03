package task

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/theo-henon/lcloud/internal/indexer"
	"github.com/theo-henon/lcloud/internal/monitoring"
	"github.com/theo-henon/lcloud/internal/plugin"
	"github.com/theo-henon/lcloud/internal/volume"
)

type Executor struct {
	macros     map[string]MacroFunc
	macroOps   *volume.MacroOps
	monitoring *monitoring.Service
	volumes    *volume.Service
	events     *plugin.Service
}

func NewExecutor(
	macroOps *volume.MacroOps,
	monitoring *monitoring.Service,
	volumes *volume.Service,
	events *plugin.Service,
) *Executor {
	return &Executor{
		macros:     DefaultRegistry(),
		macroOps:   macroOps,
		monitoring: monitoring,
		volumes:    volumes,
		events:     events,
	}
}

func (e *Executor) Run(ctx context.Context, t *Task, dryRun bool) (MacroResult, error) {
	fn, ok := e.macros[t.Macro]
	if !ok {
		return MacroResult{}, ErrInvalidMacro
	}

	params := t.Parameters
	if params == nil {
		params = Parameters{}
	}
	if dryRun {
		params = cloneParams(params)
		params["dry_run"] = true
	}

	if t.Scope == ScopeGlobal {
		return e.runGlobal(ctx, t.Macro, fn, params)
	}

	if t.VolumeID == nil {
		return MacroResult{}, ErrVolumeRequired
	}
	vol, err := e.volumes.GetByID(*t.VolumeID)
	if err != nil {
		return MacroResult{}, ErrVolumeNotFound
	}
	vc := volumeContextFrom(vol)
	return fn(ctx, e, vc, params)
}

func (e *Executor) runGlobal(ctx context.Context, macro string, fn MacroFunc, params Parameters) (MacroResult, error) {
	volumes, err := e.volumes.ListAll()
	if err != nil {
		return MacroResult{}, err
	}

	totalAffected := 0
	messages := make([]string, 0, len(volumes))
	for i := range volumes {
		if err := ctx.Err(); err != nil {
			return MacroResult{AffectedCount: totalAffected}, err
		}
		vc := volumeContextFrom(&volumes[i])
		result, err := fn(ctx, e, vc, params)
		if err != nil {
			return MacroResult{}, fmt.Errorf("volume %s: %w", volumes[i].Name, err)
		}
		totalAffected += result.AffectedCount
		if result.Message != "" {
			messages = append(messages, fmt.Sprintf("%s: %s", volumes[i].Name, result.Message))
		}
	}
	msg := fmt.Sprintf("Processed %d volume(s)", len(volumes))
	if len(messages) > 0 {
		msg = strings.Join(messages, "; ")
	}
	return MacroResult{AffectedCount: totalAffected, Message: msg}, nil
}

func volumeContextFrom(vol *volume.Volume) *VolumeContext {
	return &VolumeContext{
		ID:         vol.ID.String(),
		RootPath:   vol.RootPath,
		Name:       vol.Name,
		QuotaBytes: vol.QuotaBytes,
		UsedBytes:  vol.UsedBytes,
	}
}

func cloneParams(p Parameters) Parameters {
	out := Parameters{}
	for k, v := range p {
		out[k] = v
	}
	return out
}

func (e *Executor) volumeByContext(vc *VolumeContext) (*volume.Volume, error) {
	id, err := uuid.Parse(vc.ID)
	if err != nil {
		return nil, err
	}
	return e.volumes.GetByID(id)
}

func execDeleteOldFiles(ctx context.Context, exec *Executor, vc *VolumeContext, params Parameters) (MacroResult, error) {
	days := ParamInt(params, "days")
	cutoff := time.Now().UTC().AddDate(0, 0, -days)
	extensions := normalizeExtFilter(ParamStringSlice(params, "extensions"))
	dryRun := IsDryRun(params)

	vol, err := exec.volumeByContext(vc)
	if err != nil {
		return MacroResult{}, err
	}

	records, err := exec.macroOps.Metadata().ListAll(vol.RootPath)
	if err != nil {
		return MacroResult{}, err
	}

	affected := 0
	preview := make([]string, 0, 20)
	for _, record := range records {
		if err := ctx.Err(); err != nil {
			return MacroResult{AffectedCount: affected}, err
		}
		modified := record.ModifiedAt
		if modified.IsZero() {
			if abs, err := exec.macroOps.Paths().ResolveUserdata(vol.RootPath, record.RelativePath); err == nil {
				if info, err := os.Stat(abs); err == nil {
					modified = info.ModTime().UTC()
				}
			}
		}
		if !modified.Before(cutoff) {
			continue
		}
		if !matchesExtensions(record.Name, extensions) {
			continue
		}
		if dryRun {
			affected++
			if len(preview) < 20 {
				preview = append(preview, record.RelativePath)
			}
			continue
		}
		if err := exec.macroOps.TrashFileInternal(ctx, vol, record.RelativePath); err != nil {
			if err == volume.ErrFileNotFound {
				continue
			}
			return MacroResult{AffectedCount: affected}, err
		}
		affected++
	}

	msg := fmt.Sprintf("Trashed %d files older than %d days", affected, days)
	if dryRun {
		msg = fmt.Sprintf("Would trash %d files older than %d days", affected, days)
	}
	return MacroResult{AffectedCount: affected, Message: msg, PreviewPaths: preview}, nil
}

func execDeleteLargeFiles(ctx context.Context, exec *Executor, vc *VolumeContext, params Parameters) (MacroResult, error) {
	minBytes := int64(ParamInt(params, "min_size_mb")) * 1024 * 1024
	extensions := normalizeExtFilter(ParamStringSlice(params, "extensions"))
	dryRun := IsDryRun(params)

	vol, err := exec.volumeByContext(vc)
	if err != nil {
		return MacroResult{}, err
	}

	records, err := exec.macroOps.Metadata().ListAll(vol.RootPath)
	if err != nil {
		return MacroResult{}, err
	}

	affected := 0
	preview := make([]string, 0, 20)
	for _, record := range records {
		if err := ctx.Err(); err != nil {
			return MacroResult{AffectedCount: affected}, err
		}
		size := record.SizeBytes
		if size == 0 {
			if abs, err := exec.macroOps.Paths().ResolveUserdata(vol.RootPath, record.RelativePath); err == nil {
				if info, err := os.Stat(abs); err == nil {
					size = info.Size()
				}
			}
		}
		if size < minBytes {
			continue
		}
		if !matchesExtensions(record.Name, extensions) {
			continue
		}
		if dryRun {
			affected++
			if len(preview) < 20 {
				preview = append(preview, record.RelativePath)
			}
			continue
		}
		if err := exec.macroOps.TrashFileInternal(ctx, vol, record.RelativePath); err != nil {
			if err == volume.ErrFileNotFound {
				continue
			}
			return MacroResult{AffectedCount: affected}, err
		}
		affected++
	}

	msg := fmt.Sprintf("Trashed %d files larger than %d MB", affected, ParamInt(params, "min_size_mb"))
	if dryRun {
		msg = fmt.Sprintf("Would trash %d files larger than %d MB", affected, ParamInt(params, "min_size_mb"))
	}
	return MacroResult{AffectedCount: affected, Message: msg, PreviewPaths: preview}, nil
}

func execPurgeTrash(ctx context.Context, exec *Executor, vc *VolumeContext, params Parameters) (MacroResult, error) {
	days := ParamInt(params, "days")
	dryRun := IsDryRun(params)

	vol, err := exec.volumeByContext(vc)
	if err != nil {
		return MacroResult{}, err
	}

	affected, preview, err := exec.macroOps.PurgeTrashOlderThan(ctx, vol, days, dryRun)
	if err != nil {
		return MacroResult{}, err
	}

	msg := fmt.Sprintf("Purged %d items from trash older than %d days", affected, days)
	if dryRun {
		msg = fmt.Sprintf("Would purge %d items from trash older than %d days", affected, days)
	}
	return MacroResult{AffectedCount: affected, Message: msg, PreviewPaths: preview}, nil
}

func execClearCache(ctx context.Context, exec *Executor, vc *VolumeContext, params Parameters) (MacroResult, error) {
	dryRun := IsDryRun(params)
	vol, err := exec.volumeByContext(vc)
	if err != nil {
		return MacroResult{}, err
	}

	metaDir := filepath.Join(vol.RootPath, volume.CacheDir, volume.MetadataDir)
	thumbDir := filepath.Join(vol.RootPath, volume.CacheDir, volume.ThumbnailsDir)

	count := 0
	for _, dir := range []string{metaDir, thumbDir} {
		if err := ctx.Err(); err != nil {
			return MacroResult{AffectedCount: count}, err
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return MacroResult{}, err
		}
		for _, entry := range entries {
			if err := ctx.Err(); err != nil {
				return MacroResult{AffectedCount: count}, err
			}
			if entry.IsDir() {
				continue
			}
			count++
			if dryRun {
				continue
			}
			_ = os.Remove(filepath.Join(dir, entry.Name()))
		}
	}

	if !dryRun {
		_ = os.MkdirAll(metaDir, 0o755)
		_ = os.MkdirAll(thumbDir, 0o755)
		exec.macroOps.InvalidateStats(vol.RootPath)
	}

	msg := fmt.Sprintf("Cleared cache (%d metadata/thumbnail files)", count)
	if dryRun {
		msg = fmt.Sprintf("Would clear cache (%d metadata/thumbnail files)", count)
	}
	return MacroResult{AffectedCount: count, Message: msg}, nil
}

func execMoveFiles(ctx context.Context, exec *Executor, vc *VolumeContext, params Parameters) (MacroResult, error) {
	pattern := ParamString(params, "pattern")
	target := ParamString(params, "target_subfolder")
	dryRun := IsDryRun(params)

	vol, err := exec.volumeByContext(vc)
	if err != nil {
		return MacroResult{}, err
	}

	records, err := exec.macroOps.Metadata().ListAll(vol.RootPath)
	if err != nil {
		return MacroResult{}, err
	}

	affected := 0
	preview := make([]string, 0, 20)
	for _, record := range records {
		if err := ctx.Err(); err != nil {
			return MacroResult{AffectedCount: affected}, err
		}
		matched, err := filepath.Match(pattern, record.RelativePath)
		if err != nil || !matched {
			matched, _ = filepath.Match(pattern, record.Name)
		}
		if !matched {
			continue
		}
		dest := filepath.ToSlash(filepath.Join(target, filepath.Base(record.RelativePath)))
		if dryRun {
			affected++
			if len(preview) < 20 {
				preview = append(preview, record.RelativePath+" -> "+dest)
			}
			continue
		}
		if err := exec.macroOps.MoveFileInternal(ctx, vol, record.RelativePath, dest); err != nil {
			if err == volume.ErrDirectoryExists || err == volume.ErrFileNotFound {
				continue
			}
			return MacroResult{AffectedCount: affected}, err
		}
		affected++
	}

	msg := fmt.Sprintf("Moved %d files to %s", affected, target)
	if dryRun {
		msg = fmt.Sprintf("Would move %d files to %s", affected, target)
	}
	return MacroResult{AffectedCount: affected, Message: msg, PreviewPaths: preview}, nil
}

func execSortByType(ctx context.Context, exec *Executor, vc *VolumeContext, params Parameters) (MacroResult, error) {
	dryRun := IsDryRun(params)
	vol, err := exec.volumeByContext(vc)
	if err != nil {
		return MacroResult{}, err
	}

	rootFiles, err := listUserdataRootFiles(exec, vol)
	if err != nil {
		return MacroResult{}, err
	}

	affected := 0
	preview := make([]string, 0, 20)
	for _, record := range rootFiles {
		if err := ctx.Err(); err != nil {
			return MacroResult{AffectedCount: affected}, err
		}
		category := monitoring.CategoryForMIME(record.MimeType)
		folder := sortFolderForCategory(category)
		dest := filepath.ToSlash(filepath.Join(folder, record.Name))
		if dryRun {
			affected++
			if len(preview) < 20 {
				preview = append(preview, record.RelativePath+" -> "+dest)
			}
			continue
		}
		if err := exec.macroOps.MoveFileInternal(ctx, vol, record.RelativePath, dest); err != nil {
			if err == volume.ErrDirectoryExists || err == volume.ErrFileNotFound {
				continue
			}
			return MacroResult{AffectedCount: affected}, err
		}
		affected++
	}

	msg := fmt.Sprintf("Sorted %d files by type", affected)
	if dryRun {
		msg = fmt.Sprintf("Would sort %d files by type", affected)
	}
	return MacroResult{AffectedCount: affected, Message: msg, PreviewPaths: preview}, nil
}

func execSortByDate(ctx context.Context, exec *Executor, vc *VolumeContext, params Parameters) (MacroResult, error) {
	dryRun := IsDryRun(params)
	vol, err := exec.volumeByContext(vc)
	if err != nil {
		return MacroResult{}, err
	}

	rootFiles, err := listUserdataRootFiles(exec, vol)
	if err != nil {
		return MacroResult{}, err
	}

	affected := 0
	preview := make([]string, 0, 20)
	for _, record := range rootFiles {
		if err := ctx.Err(); err != nil {
			return MacroResult{AffectedCount: affected}, err
		}
		modified := record.ModifiedAt
		if modified.IsZero() {
			modified = time.Now().UTC()
		}
		folder := fmt.Sprintf("%04d/%02d", modified.Year(), int(modified.Month()))
		dest := filepath.ToSlash(filepath.Join(folder, record.Name))
		if dryRun {
			affected++
			if len(preview) < 20 {
				preview = append(preview, record.RelativePath+" -> "+dest)
			}
			continue
		}
		if err := exec.macroOps.MoveFileInternal(ctx, vol, record.RelativePath, dest); err != nil {
			if err == volume.ErrDirectoryExists || err == volume.ErrFileNotFound {
				continue
			}
			return MacroResult{AffectedCount: affected}, err
		}
		affected++
	}

	msg := fmt.Sprintf("Sorted %d files by date", affected)
	if dryRun {
		msg = fmt.Sprintf("Would sort %d files by date", affected)
	}
	return MacroResult{AffectedCount: affected, Message: msg, PreviewPaths: preview}, nil
}

func execRebuildIndex(ctx context.Context, exec *Executor, vc *VolumeContext, params Parameters) (MacroResult, error) {
	_ = params
	vol, err := exec.volumeByContext(vc)
	if err != nil {
		return MacroResult{}, err
	}

	idx, err := exec.macroOps.IndexManager().Get(vol.RootPath)
	if err != nil {
		return MacroResult{}, err
	}
	if err := idx.Rebuild(vol.RootPath); err != nil {
		return MacroResult{}, err
	}

	records, err := exec.macroOps.Metadata().ListAll(vol.RootPath)
	if err != nil {
		return MacroResult{}, err
	}
	for _, record := range records {
		if err := ctx.Err(); err != nil {
			return MacroResult{AffectedCount: 0}, err
		}
		if err := idx.Index(indexer.FileMetadata{
			ID:           record.ID,
			Name:         record.Name,
			RelativePath: record.RelativePath,
			MimeType:     record.MimeType,
			SizeBytes:    record.SizeBytes,
			ModifiedAt:   record.ModifiedAt,
			SHA256:       record.SHA256,
		}); err != nil {
			return MacroResult{AffectedCount: 0}, err
		}
	}

	return MacroResult{
		AffectedCount: len(records),
		Message:       fmt.Sprintf("Rebuilt index with %d files", len(records)),
	}, nil
}

func execComputeStats(ctx context.Context, exec *Executor, vc *VolumeContext, params Parameters) (MacroResult, error) {
	_ = ctx
	_ = params
	vol, err := exec.volumeByContext(vc)
	if err != nil {
		return MacroResult{}, err
	}
	stats, err := exec.monitoring.ComputeStats(vol)
	if err != nil {
		return MacroResult{}, err
	}
	return MacroResult{
		AffectedCount: stats.FileCount,
		Message:       fmt.Sprintf("Computed stats for %d files", stats.FileCount),
	}, nil
}

func execAlertUsage(ctx context.Context, exec *Executor, vc *VolumeContext, params Parameters) (MacroResult, error) {
	_ = ctx
	threshold := ParamInt(params, "threshold_percent")
	vol, err := exec.volumeByContext(vc)
	if err != nil {
		return MacroResult{}, err
	}
	if vol.QuotaBytes <= 0 {
		return MacroResult{Message: "Volume has no quota configured"}, nil
	}

	usagePercent := monitoring.UsagePercentInt(vol.UsedBytes, vol.QuotaBytes)
	if usagePercent < threshold {
		return MacroResult{
			Message: fmt.Sprintf("Usage at %d%%, below threshold %d%%", usagePercent, threshold),
		}, nil
	}

	if exec.events != nil {
		exec.events.PublishEvent(context.Background(), plugin.FromVolumeAlertUsage(
			vol.ID.String(), usagePercent, vol.QuotaBytes, vol.UsedBytes,
		))
	}

	return MacroResult{
		AffectedCount: 1,
		Message:       fmt.Sprintf("Alert: usage at %d%% (threshold %d%%)", usagePercent, threshold),
	}, nil
}

func listUserdataRootFiles(exec *Executor, vol *volume.Volume) ([]volume.FileMetadataRecord, error) {
	records, err := exec.macroOps.Metadata().ListAll(vol.RootPath)
	if err != nil {
		return nil, err
	}
	root := make([]volume.FileMetadataRecord, 0)
	for _, record := range records {
		if !isUserdataRootFile(record.RelativePath) {
			continue
		}
		root = append(root, record)
	}
	return root, nil
}

func isUserdataRootFile(relPath string) bool {
	relPath = filepath.ToSlash(strings.TrimSpace(relPath))
	if relPath == "" || relPath == "." {
		return false
	}
	return !strings.Contains(relPath, "/")
}

func sortFolderForCategory(category string) string {
	switch category {
	case monitoring.CategoryImages:
		return "images"
	case monitoring.CategoryDocuments:
		return "documents"
	case monitoring.CategoryVideos:
		return "videos"
	case monitoring.CategoryAudio:
		return "audio"
	default:
		return "other"
	}
}

func normalizeExtFilter(extensions []string) []string {
	if len(extensions) == 0 {
		return nil
	}
	return volume.NormalizeExtensions(extensions)
}

func matchesExtensions(filename string, extensions []string) bool {
	if len(extensions) == 0 {
		return true
	}
	ext := volume.NormalizeExtension(filepath.Ext(filename))
	for _, item := range extensions {
		if ext == item {
			return true
		}
	}
	return false
}
