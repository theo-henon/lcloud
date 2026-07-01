package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/theo-henon/lcloud/internal/volume"
	"github.com/theo-henon/lcloud/pkg/httputil"
)

type DiskHandler struct {
	disks *volume.DiskRegistry
}

func NewDiskHandler(disks *volume.DiskRegistry) *DiskHandler {
	return &DiskHandler{disks: disks}
}

func (h *DiskHandler) List(c *gin.Context) {
	httputil.JSON(c, http.StatusOK, gin.H{"disks": h.disks.ListDisks()})
}
