package volume

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateExtensionAllowMode(t *testing.T) {
	filters := Filters{
		Mode:       FilterModeAllow,
		Extensions: []string{".jpg", ".png"},
	}
	require.NoError(t, ValidateExtension(filters, "photo.JPG"))
	require.ErrorIs(t, ValidateExtension(filters, "video.mp4"), ErrFileFilterRejected)
}

func TestValidateExtensionBlockMode(t *testing.T) {
	filters := Filters{
		Mode:       FilterModeBlock,
		Extensions: []string{".exe"},
	}
	require.NoError(t, ValidateExtension(filters, "notes.txt"))
	require.ErrorIs(t, ValidateExtension(filters, "setup.exe"), ErrFileFilterRejected)
}

func TestValidateFilters(t *testing.T) {
	require.NoError(t, ValidateFilters(Filters{}))
	require.ErrorIs(t, ValidateFilters(Filters{Mode: "invalid"}), ErrInvalidFilter)
	require.NoError(t, ValidateFilters(Filters{Mode: FilterModeAllow}))
}

func TestNormalizeFiltersClearsModeWithoutExtensions(t *testing.T) {
	normalized := NormalizeFilters(Filters{
		Mode:       FilterModeAllow,
		Extensions: []string{},
	})
	require.Equal(t, FilterMode(""), normalized.Mode)
	require.Empty(t, normalized.Extensions)
}

func TestNormalizeExtensions(t *testing.T) {
	require.Equal(t, []string{".jpg", ".png"}, NormalizeExtensions([]string{"jpg", ".PNG", ".jpg"}))
}
