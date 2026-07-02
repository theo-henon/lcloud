package task

import (
	"context"
)

type MacroFunc func(ctx context.Context, exec *Executor, vol *VolumeContext, params Parameters) (MacroResult, error)

type VolumeContext struct {
	ID       string
	RootPath string
	Name     string
	QuotaBytes int64
	UsedBytes  int64
}

func DefaultRegistry() map[string]MacroFunc {
	return map[string]MacroFunc{
		MacroDeleteOldFiles:   execDeleteOldFiles,
		MacroDeleteLargeFiles: execDeleteLargeFiles,
		MacroClearCache:       execClearCache,
		MacroMoveFiles:        execMoveFiles,
		MacroSortByType:       execSortByType,
		MacroSortByDate:       execSortByDate,
		MacroRebuildIndex:     execRebuildIndex,
		MacroComputeStats:     execComputeStats,
		MacroAlertUsage:       execAlertUsage,
	}
}
