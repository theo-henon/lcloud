package indexer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBleveIndexerSearchFilters(t *testing.T) {
	root := t.TempDir()
	idx, err := OpenBleveIndexer(root)
	require.NoError(t, err)
	t.Cleanup(func() { _ = idx.Close() })

	now := time.Now().UTC()
	files := []FileMetadata{
		{ID: "1", Name: "photo.jpg", RelativePath: "photo.jpg", MimeType: "image/jpeg", SizeBytes: 2048, ModifiedAt: now, SHA256: "a"},
		{ID: "2", Name: "doc.pdf", RelativePath: "doc.pdf", MimeType: "application/pdf", SizeBytes: 512, ModifiedAt: now.Add(-48 * time.Hour), SHA256: "b"},
		{ID: "3", Name: "large.png", RelativePath: "large.png", MimeType: "image/png", SizeBytes: 4096, ModifiedAt: now, SHA256: "c"},
	}
	for _, file := range files {
		require.NoError(t, idx.Index(file))
	}

	mimePrefix := "image/"
	results, err := idx.Search(SearchQuery{MimePrefix: mimePrefix})
	require.NoError(t, err)
	require.Len(t, results, 2)

	minSize := int64(3000)
	results, err = idx.Search(SearchQuery{MinSizeBytes: &minSize})
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "large.png", results[0].Name)

	modifiedAfter := now.Add(-24 * time.Hour)
	results, err = idx.Search(SearchQuery{ModifiedAfter: &modifiedAfter, MimePrefix: mimePrefix})
	require.NoError(t, err)
	require.Len(t, results, 2)
}
