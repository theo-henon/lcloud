package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/settings"
	"github.com/theo-henon/lcloud/internal/volume"
	"github.com/theo-henon/lcloud/pkg/httputil"
)

type DiskHandler struct {
	disks     *volume.DiskRegistry
	settings  *settings.Service
}

func NewDiskHandler(disks *volume.DiskRegistry, settingsService *settings.Service) *DiskHandler {
	return &DiskHandler{disks: disks, settings: settingsService}
}

func (h *DiskHandler) List(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httputil.Unauthorized(c, "unauthorized")
		return
	}

	disks := h.disks.ListDisks()
	// Admins need real disk identifiers when creating volumes.
	mask := false
	if claims.Role != auth.RoleAdmin {
		var err error
		mask, err = h.settings.ShouldMaskFor(claims)
		if err != nil {
			httputil.InternalError(c, "unable to load settings")
			return
		}
	}
	if mask {
		masked := make([]volume.DiskInfo, len(disks))
		for i, disk := range disks {
			masked[i] = maskDiskInfo(disk, i)
		}
		disks = masked
	}

	httputil.JSON(c, http.StatusOK, gin.H{"disks": disks})
}
