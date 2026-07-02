package task

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestValidateParametersDeleteOldFiles(t *testing.T) {
	require.NoError(t, ValidateParameters(MacroDeleteOldFiles, Parameters{"days": 90}))
	require.Error(t, ValidateParameters(MacroDeleteOldFiles, Parameters{}))
	require.Error(t, ValidateParameters(MacroComputeStats, Parameters{"dry_run": true}))
}

func TestValidateMacroScope(t *testing.T) {
	require.NoError(t, ValidateMacroScope(MacroDeleteOldFiles, ScopeVolume))
	require.Error(t, ValidateMacroScope(MacroDeleteOldFiles, ScopeGlobal))
	require.NoError(t, ValidateMacroScope(MacroComputeStats, ScopeGlobal))
}

func TestValidateSchedule(t *testing.T) {
	require.NoError(t, ValidateSchedule(ScheduleTypeCron, "0 2 * * 0"))
	require.NoError(t, ValidateSchedule(ScheduleTypeInterval, "1h"))
	require.Error(t, ValidateSchedule(ScheduleTypeInterval, "30s"))
	require.Error(t, ValidateSchedule(ScheduleTypeCron, "not a cron"))
}

func TestNextRunAt(t *testing.T) {
	from := timeMustParse("2026-07-01T10:00:00Z")
	next, err := NextRunAt(ScheduleTypeInterval, "1h", from)
	require.NoError(t, err)
	require.Equal(t, "2026-07-01T11:00:00Z", next.UTC().Format(timeRFC))
}

func timeMustParse(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

const timeRFC = "2006-01-02T15:04:05Z"
