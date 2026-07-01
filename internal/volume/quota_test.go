package volume

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestQuotaAllowsUnlimited(t *testing.T) {
	vol := &Volume{QuotaBytes: 0, UsedBytes: 1000}
	require.NoError(t, QuotaAllows(vol, 1_000_000))
}

func TestQuotaExceeded(t *testing.T) {
	vol := &Volume{QuotaBytes: 100, UsedBytes: 90}
	require.ErrorIs(t, QuotaAllows(vol, 20), ErrQuotaExceeded)
	require.NoError(t, QuotaAllows(vol, 10))
}

func TestApplyUsageDelta(t *testing.T) {
	vol := &Volume{UsedBytes: 10}
	ApplyUsageDelta(vol, 5)
	require.Equal(t, int64(15), vol.UsedBytes)
	ApplyUsageDelta(vol, -20)
	require.Equal(t, int64(0), vol.UsedBytes)
}

func TestVolumeFiltersJSON(t *testing.T) {
	vol := Volume{
		ID:      uuid.New(),
		Filters: Filters{Mode: FilterModeAllow, Extensions: []string{".jpg"}},
	}
	value, err := vol.Filters.Value()
	require.NoError(t, err)

	var scanned Filters
	require.NoError(t, scanned.Scan(value))
	require.Equal(t, vol.Filters, scanned)
}
