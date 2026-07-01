package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/indexer"
	"github.com/theo-henon/lcloud/internal/volume"
	"github.com/theo-henon/lcloud/pkg/httputil"
)

type SearchHandler struct {
	volumes      *volume.Service
	indexManager *indexer.IndexManager
	metadata     *volume.MetadataCache
}

func NewSearchHandler(volumes *volume.Service, indexManager *indexer.IndexManager) *SearchHandler {
	return &SearchHandler{
		volumes:      volumes,
		indexManager: indexManager,
		metadata:     volume.NewMetadataCache(),
	}
}

type searchResultItem struct {
	Name          string    `json:"name"`
	RelativePath  string    `json:"relative_path"`
	MimeType      string    `json:"mime_type"`
	SizeBytes     int64     `json:"size_bytes"`
	ModifiedAt    time.Time `json:"modified_at"`
	HasThumbnail  bool      `json:"has_thumbnail"`
}

type searchResponse struct {
	Query   map[string]any     `json:"query"`
	Total   int                `json:"total"`
	Results []searchResultItem `json:"results"`
}

func (h *SearchHandler) Search(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httputil.Unauthorized(c, "unauthorized")
		return
	}

	volumeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.BadRequest(c, "invalid volume id")
		return
	}

	vol, err := h.volumes.Get(claims, volumeID)
	if err != nil {
		if errors.Is(err, volume.ErrVolumeNotFound) {
			httputil.Unprocessable(c, "VOLUME_NOT_FOUND", "volume not found")
			return
		}
		if errors.Is(err, volume.ErrForbidden) {
			httputil.Forbidden(c, "forbidden")
			return
		}
		httputil.InternalError(c, "unable to load volume")
		return
	}

	query, queryMeta, err := parseSearchQuery(c)
	if err != nil {
		httputil.Unprocessable(c, "INVALID_SEARCH_QUERY", err.Error())
		return
	}

	idx, err := h.indexManager.Get(vol.RootPath)
	if err != nil {
		httputil.Unprocessable(c, "INDEX_UNAVAILABLE", "search index unavailable")
		return
	}

	results, err := idx.Search(query)
	if err != nil {
		httputil.InternalError(c, "unable to search volume")
		return
	}

	items := make([]searchResultItem, 0, len(results))
	for _, result := range results {
		item := searchResultItem{
			Name:         result.Name,
			RelativePath: result.RelativePath,
			MimeType:     result.MimeType,
			SizeBytes:    result.SizeBytes,
			ModifiedAt:   result.ModifiedAt,
		}
		if record, err := h.metadata.ReadByRelativePath(vol.RootPath, result.RelativePath); err == nil {
			item.HasThumbnail = record.ThumbnailPath != ""
		}
		items = append(items, item)
	}

	httputil.JSON(c, http.StatusOK, searchResponse{
		Query:   queryMeta,
		Total:   len(items),
		Results: items,
	})
}

func parseSearchQuery(c *gin.Context) (indexer.SearchQuery, map[string]any, error) {
	query := indexer.SearchQuery{
		Term:       strings.TrimSpace(c.Query("q")),
		MimePrefix: strings.TrimSpace(c.Query("mime_prefix")),
	}

	if raw := strings.TrimSpace(c.Query("min_size")); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value < 0 {
			return query, nil, errors.New("min_size must be a non-negative integer")
		}
		query.MinSizeBytes = &value
	}
	if raw := strings.TrimSpace(c.Query("max_size")); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value < 0 {
			return query, nil, errors.New("max_size must be a non-negative integer")
		}
		query.MaxSizeBytes = &value
	}
	if raw := strings.TrimSpace(c.Query("modified_after")); raw != "" {
		value, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return query, nil, errors.New("modified_after must be RFC3339")
		}
		query.ModifiedAfter = &value
	}
	if raw := strings.TrimSpace(c.Query("modified_before")); raw != "" {
		value, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return query, nil, errors.New("modified_before must be RFC3339")
		}
		query.ModifiedBefore = &value
	}
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 {
			return query, nil, errors.New("limit must be a positive integer")
		}
		query.Limit = value
	}

	meta := map[string]any{
		"q":           query.Term,
		"mime_prefix": query.MimePrefix,
		"limit":       query.Limit,
	}
	if query.MinSizeBytes != nil {
		meta["min_size"] = *query.MinSizeBytes
	}
	if query.MaxSizeBytes != nil {
		meta["max_size"] = *query.MaxSizeBytes
	}
	if query.ModifiedAfter != nil {
		meta["modified_after"] = query.ModifiedAfter.Format(time.RFC3339)
	}
	if query.ModifiedBefore != nil {
		meta["modified_before"] = query.ModifiedBefore.Format(time.RFC3339)
	}
	if query.Limit == 0 {
		meta["limit"] = 50
	}

	return query, meta, nil
}
