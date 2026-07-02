package task

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParametersJSONRoundTrip(t *testing.T) {
	params := Parameters{"days": float64(90), "dry_run": true}
	val, err := params.Value()
	require.NoError(t, err)

	var scanned Parameters
	require.NoError(t, scanned.Scan(val))
	require.Equal(t, float64(90), scanned["days"])
	require.Equal(t, true, scanned["dry_run"])
}

func TestRunStatusConstants(t *testing.T) {
	require.Equal(t, "dry_run", RunStatusDryRun)
	require.Equal(t, "skipped", RunStatusSkipped)
}
