package volume

func QuotaAllows(vol *Volume, additionalBytes int64) error {
	if vol.QuotaBytes == 0 {
		return nil
	}
	if vol.UsedBytes+additionalBytes > vol.QuotaBytes {
		return ErrQuotaExceeded
	}
	return nil
}

func ApplyUsageDelta(vol *Volume, delta int64) {
	vol.UsedBytes += delta
	if vol.UsedBytes < 0 {
		vol.UsedBytes = 0
	}
}
