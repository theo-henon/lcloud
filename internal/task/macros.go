package task

import (
	"fmt"
	"strings"
)

const (
	MacroDeleteOldFiles   = "delete_old_files"
	MacroDeleteLargeFiles = "delete_large_files"
	MacroClearCache       = "clear_cache"
	MacroMoveFiles        = "move_files"
	MacroSortByType       = "sort_by_type"
	MacroSortByDate       = "sort_by_date"
	MacroRebuildIndex     = "rebuild_index"
	MacroComputeStats     = "compute_stats"
	MacroAlertUsage       = "alert_usage"
	MacroPurgeTrash       = "purge_trash"
)

var allMacros = map[string]struct{}{
	MacroDeleteOldFiles:   {},
	MacroDeleteLargeFiles: {},
	MacroClearCache:       {},
	MacroMoveFiles:        {},
	MacroSortByType:       {},
	MacroSortByDate:       {},
	MacroRebuildIndex:     {},
	MacroComputeStats:     {},
	MacroAlertUsage:       {},
	MacroPurgeTrash:       {},
}

var destructiveMacros = map[string]struct{}{
	MacroDeleteOldFiles:   {},
	MacroDeleteLargeFiles: {},
	MacroClearCache:       {},
	MacroMoveFiles:        {},
	MacroSortByType:       {},
	MacroSortByDate:       {},
	MacroPurgeTrash:       {},
}

var globalAllowedMacros = map[string]struct{}{
	MacroRebuildIndex: {},
	MacroComputeStats: {},
	MacroAlertUsage:   {},
}

func IsValidMacro(name string) bool {
	_, ok := allMacros[name]
	return ok
}

func IsDestructiveMacro(name string) bool {
	_, ok := destructiveMacros[name]
	return ok
}

func AllowsGlobalScope(name string) bool {
	_, ok := globalAllowedMacros[name]
	return ok
}

func ValidateMacroScope(macro, scope string) error {
	if !IsValidMacro(macro) {
		return ErrInvalidMacro
	}
	if scope == ScopeGlobal && !AllowsGlobalScope(macro) {
		return fmt.Errorf("%w: macro %s cannot be global", ErrInvalidMacro, macro)
	}
	if scope != ScopeVolume && scope != ScopeGlobal {
		return fmt.Errorf("%w: unknown scope %s", ErrInvalidParameters, scope)
	}
	return nil
}

func ValidateParameters(macro string, params Parameters) error {
	if params == nil {
		params = Parameters{}
	}
	if dryRun, ok := params["dry_run"]; ok {
		if _, isDestructive := destructiveMacros[macro]; !isDestructive && dryRun != nil {
			return fmt.Errorf("%w: dry_run not supported for macro %s", ErrInvalidParameters, macro)
		}
	}

	switch macro {
	case MacroDeleteOldFiles:
		return validateDeleteOldFilesParams(params)
	case MacroDeleteLargeFiles:
		return validateDeleteLargeFilesParams(params)
	case MacroClearCache:
		return validateOptionalDryRun(params)
	case MacroMoveFiles:
		return validateMoveFilesParams(params)
	case MacroSortByType, MacroSortByDate:
		return validateOptionalDryRun(params)
	case MacroRebuildIndex, MacroComputeStats:
		return nil
	case MacroAlertUsage:
		return validateAlertUsageParams(params)
	case MacroPurgeTrash:
		return validateDeleteOldFilesParams(params)
	default:
		return ErrInvalidMacro
	}
}

func validateOptionalDryRun(params Parameters) error {
	if dryRun, ok := params["dry_run"]; ok && dryRun != nil {
		if _, ok := dryRun.(bool); !ok {
			return fmt.Errorf("%w: dry_run must be boolean", ErrInvalidParameters)
		}
	}
	return nil
}

func validateDeleteOldFilesParams(params Parameters) error {
	if err := validateOptionalDryRun(params); err != nil {
		return err
	}
	days, ok := params["days"]
	if !ok {
		return fmt.Errorf("%w: days is required", ErrInvalidParameters)
	}
	switch v := days.(type) {
	case float64:
		if v < 1 {
			return fmt.Errorf("%w: days must be >= 1", ErrInvalidParameters)
		}
	case int:
		if v < 1 {
			return fmt.Errorf("%w: days must be >= 1", ErrInvalidParameters)
		}
	default:
		return fmt.Errorf("%w: days must be a number", ErrInvalidParameters)
	}
	return validateExtensions(params)
}

func validateDeleteLargeFilesParams(params Parameters) error {
	if err := validateOptionalDryRun(params); err != nil {
		return err
	}
	size, ok := params["min_size_mb"]
	if !ok {
		return fmt.Errorf("%w: min_size_mb is required", ErrInvalidParameters)
	}
	switch v := size.(type) {
	case float64:
		if v < 1 {
			return fmt.Errorf("%w: min_size_mb must be >= 1", ErrInvalidParameters)
		}
	case int:
		if v < 1 {
			return fmt.Errorf("%w: min_size_mb must be >= 1", ErrInvalidParameters)
		}
	default:
		return fmt.Errorf("%w: min_size_mb must be a number", ErrInvalidParameters)
	}
	return validateExtensions(params)
}

func validateMoveFilesParams(params Parameters) error {
	if err := validateOptionalDryRun(params); err != nil {
		return err
	}
	pattern, _ := params["pattern"].(string)
	if strings.TrimSpace(pattern) == "" {
		return fmt.Errorf("%w: pattern is required", ErrInvalidParameters)
	}
	target, _ := params["target_subfolder"].(string)
	if strings.TrimSpace(target) == "" {
		return fmt.Errorf("%w: target_subfolder is required", ErrInvalidParameters)
	}
	return nil
}

func validateAlertUsageParams(params Parameters) error {
	threshold, ok := params["threshold_percent"]
	if !ok {
		return fmt.Errorf("%w: threshold_percent is required", ErrInvalidParameters)
	}
	switch v := threshold.(type) {
	case float64:
		if v < 1 || v > 100 {
			return fmt.Errorf("%w: threshold_percent must be 1-100", ErrInvalidParameters)
		}
	case int:
		if v < 1 || v > 100 {
			return fmt.Errorf("%w: threshold_percent must be 1-100", ErrInvalidParameters)
		}
	default:
		return fmt.Errorf("%w: threshold_percent must be a number", ErrInvalidParameters)
	}
	return nil
}

func validateExtensions(params Parameters) error {
	extRaw, ok := params["extensions"]
	if !ok || extRaw == nil {
		return nil
	}
	list, ok := extRaw.([]any)
	if !ok {
		return fmt.Errorf("%w: extensions must be an array", ErrInvalidParameters)
	}
	for _, item := range list {
		if _, ok := item.(string); !ok {
			return fmt.Errorf("%w: extensions must be strings", ErrInvalidParameters)
		}
	}
	return nil
}

func ParamBool(params Parameters, key string) bool {
	v, ok := params[key]
	if !ok || v == nil {
		return false
	}
	b, _ := v.(bool)
	return b
}

func ParamInt(params Parameters, key string) int {
	v, ok := params[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	default:
		return 0
	}
}

func ParamString(params Parameters, key string) string {
	v, _ := params[key].(string)
	return strings.TrimSpace(v)
}

func ParamStringSlice(params Parameters, key string) []string {
	raw, ok := params[key]
	if !ok || raw == nil {
		return nil
	}
	list, ok := raw.([]any)
	if !ok {
		if s, ok := raw.([]string); ok {
			return s
		}
		return nil
	}
	out := make([]string, 0, len(list))
	for _, item := range list {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func IsDryRun(params Parameters) bool {
	return ParamBool(params, "dry_run")
}
