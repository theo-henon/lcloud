package monitoring

import "testing"

func TestUsagePercentInt(t *testing.T) {
	const gb = 1024 * 1024 * 1024
	quota := int64(50 * gb)

	// ~1% usage — integer division truncates to 0; rounded percent must be 1.
	used := quota / 100
	if got := UsagePercentInt(used, quota); got != 1 {
		t.Fatalf("UsagePercentInt(~1%%) = %d, want 1", got)
	}

	if got := UsagePercentInt(0, quota); got != 0 {
		t.Fatalf("UsagePercentInt(0) = %d, want 0", got)
	}

	if got := UsagePercentInt(quota, quota); got != 100 {
		t.Fatalf("UsagePercentInt(full) = %d, want 100", got)
	}
}
