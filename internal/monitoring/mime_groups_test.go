package monitoring

import "testing"

func TestCategoryForMIME(t *testing.T) {
	tests := []struct {
		mime     string
		expected string
	}{
		{"image/jpeg", CategoryImages},
		{"video/mp4", CategoryVideos},
		{"audio/mpeg", CategoryAudio},
		{"application/pdf", CategoryDocuments},
		{"text/plain", CategoryDocuments},
		{"application/vnd.openxmlformats-officedocument.wordprocessingml.document", CategoryDocuments},
		{"application/zip", CategoryArchives},
		{"application/x-tar", CategoryArchives},
		{"application/octet-stream", CategoryOther},
		{"", CategoryOther},
	}

	for _, tc := range tests {
		got := CategoryForMIME(tc.mime)
		if got != tc.expected {
			t.Errorf("CategoryForMIME(%q) = %q, want %q", tc.mime, got, tc.expected)
		}
	}
}
