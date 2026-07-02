package api

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/volume"
	"github.com/theo-henon/lcloud/pkg/httputil"
)

type FileHandler struct {
	files *volume.FileService
}

func NewFileHandler(files *volume.FileService) *FileHandler {
	return &FileHandler{files: files}
}

type createDirectoryRequest struct {
	Path string `json:"path" binding:"required"`
}

func (h *FileHandler) List(c *gin.Context) {
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

	listing, err := h.files.List(claims, volumeID, c.DefaultQuery("path", "."))
	if err != nil {
		mapFileError(c, err)
		return
	}
	httputil.JSON(c, http.StatusOK, listing)
}

func (h *FileHandler) CreateDirectory(c *gin.Context) {
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

	var req createDirectoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "invalid request body")
		return
	}

	if err := h.files.CreateDirectory(claims, volumeID, req.Path); err != nil {
		mapFileError(c, err)
		return
	}
	httputil.JSON(c, http.StatusCreated, gin.H{"path": req.Path})
}

func (h *FileHandler) Upload(c *gin.Context) {
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

	fileHeader, err := c.FormFile("file")
	if err != nil {
		httputil.BadRequest(c, "file is required")
		return
	}

	src, err := fileHeader.Open()
	if err != nil {
		httputil.InternalError(c, "unable to read upload")
		return
	}
	defer src.Close()

	entry, err := h.files.Upload(claims, volumeID, c.PostForm("path"), fileHeader.Filename, src, fileHeader.Size)
	if err != nil {
		mapFileError(c, err)
		return
	}
	httputil.JSON(c, http.StatusCreated, entry)
}

type moveFileRequest struct {
	FromPath string `json:"from_path" binding:"required"`
	ToPath   string `json:"to_path" binding:"required"`
}

type renameFileRequest struct {
	Path    string `json:"path" binding:"required"`
	NewName string `json:"new_name" binding:"required"`
}

func (h *FileHandler) Move(c *gin.Context) {
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

	var req moveFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "invalid request body")
		return
	}

	entry, err := h.files.Move(claims, volumeID, req.FromPath, req.ToPath)
	if err != nil {
		mapFileError(c, err)
		return
	}
	httputil.JSON(c, http.StatusOK, entry)
}

func (h *FileHandler) Rename(c *gin.Context) {
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

	var req renameFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "invalid request body")
		return
	}

	entry, err := h.files.Rename(claims, volumeID, req.Path, req.NewName)
	if err != nil {
		mapFileError(c, err)
		return
	}
	httputil.JSON(c, http.StatusOK, entry)
}

func (h *FileHandler) Download(c *gin.Context) {
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

	relPath := c.Query("path")
	if relPath == "" {
		httputil.BadRequest(c, "path is required")
		return
	}

	file, mimeType, err := h.files.OpenContent(claims, volumeID, relPath)
	if err != nil {
		mapFileError(c, err)
		return
	}
	defer file.Close()

	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	disposition := c.DefaultQuery("disposition", "attachment")
	c.Header("Content-Type", mimeType)
	if disposition == "inline" {
		c.Header("Content-Disposition", httputil.ContentDispositionInline(relPath))
	} else {
		c.Header("Content-Disposition", httputil.ContentDispositionAttachment(relPath))
	}
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, file)
}

func (h *FileHandler) Thumbnail(c *gin.Context) {
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

	relPath := c.Query("path")
	if relPath == "" {
		httputil.BadRequest(c, "path is required")
		return
	}

	file, err := h.files.OpenThumbnail(claims, volumeID, relPath)
	if err != nil {
		mapFileError(c, err)
		return
	}
	defer file.Close()

	c.Header("Content-Type", "image/jpeg")
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, file)
}

func (h *FileHandler) Delete(c *gin.Context) {
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

	relPath := c.Query("path")
	if relPath == "" {
		httputil.BadRequest(c, "path is required")
		return
	}

	if err := h.files.Delete(claims, volumeID, relPath); err != nil {
		mapFileError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func mapFileError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, volume.ErrVolumeNotFound), errors.Is(err, volume.ErrFileNotFound):
		httputil.Error(c, http.StatusNotFound, "NOT_FOUND", "resource not found")
	case errors.Is(err, volume.ErrForbidden):
		httputil.Forbidden(c, "forbidden")
	case errors.Is(err, volume.ErrPathTraversal):
		httputil.Unprocessable(c, "PATH_TRAVERSAL", "invalid path")
	case errors.Is(err, volume.ErrFileFilterRejected):
		httputil.Unprocessable(c, "FILE_FILTER_REJECTED", "file extension not allowed")
	case errors.Is(err, volume.ErrQuotaExceeded):
		httputil.Unprocessable(c, "QUOTA_EXCEEDED", "quota exceeded")
	case errors.Is(err, volume.ErrUploadTooLarge):
		httputil.Error(c, http.StatusRequestEntityTooLarge, "UPLOAD_TOO_LARGE", "upload exceeds max size")
	case errors.Is(err, volume.ErrUploadSizeMismatch):
		httputil.BadRequest(c, "upload size mismatch")
	case errors.Is(err, volume.ErrDirectoryExists):
		httputil.Error(c, http.StatusConflict, "PATH_EXISTS", "path already exists")
	case errors.Is(err, volume.ErrInvalidEntryName):
		httputil.Unprocessable(c, "INVALID_PATH", "invalid path")
	case errors.Is(err, volume.ErrNotAFile):
		httputil.Unprocessable(c, "NOT_A_FILE", "not a file")
	case errors.Is(err, volume.ErrNotDirectory):
		httputil.BadRequest(c, "not a directory")
	default:
		httputil.InternalError(c, "file operation failed")
	}
}
