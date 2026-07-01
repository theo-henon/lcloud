package indexer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBleveIndexerIndexSearchDelete(t *testing.T) {
	root := t.TempDir()

	idx, err := OpenBleveIndexer(root)
	require.NoError(t, err)
	t.Cleanup(func() { _ = idx.Close() })

	count, err := idx.index.DocCount()
	require.NoError(t, err)
	require.Equal(t, uint64(0), count)

	meta := FileMetadata{
		ID:           "file-1",
		Name:         "vacation.jpg",
		RelativePath: "vacation.jpg",
		MimeType:     "image/jpeg",
		SizeBytes:    1024,
		ModifiedAt:   time.Now().UTC(),
		SHA256:       "abc123",
	}
	require.NoError(t, idx.Index(meta))

	count, err = idx.index.DocCount()
	require.NoError(t, err)
	require.Equal(t, uint64(1), count)

	require.NoError(t, idx.Delete(meta.RelativePath))
	count, err = idx.index.DocCount()
	require.NoError(t, err)
	require.Equal(t, uint64(0), count)
}
