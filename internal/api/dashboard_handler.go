package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/dashboard"
	"github.com/theo-henon/lcloud/pkg/httputil"
)

type DashboardHandler struct {
	service *dashboard.Service
}

func NewDashboardHandler(service *dashboard.Service) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (h *DashboardHandler) GetMe(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httputil.Unauthorized(c, "unauthorized")
		return
	}

	result, err := h.service.GetLayout(c.Request.Context(), claims)
	if err != nil {
		httputil.InternalError(c, "unable to load dashboard layout")
		return
	}

	httputil.JSON(c, http.StatusOK, result)
}

func (h *DashboardHandler) PatchMe(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httputil.Unauthorized(c, "unauthorized")
		return
	}

	var input dashboard.PatchLayoutInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.BadRequest(c, "invalid request body")
		return
	}

	result, err := h.service.SaveLayout(c.Request.Context(), claims, input.Layout)
	if err != nil {
		switch {
		case errors.Is(err, dashboard.ErrUnknownWidget):
			httputil.Error(c, http.StatusBadRequest, "UNKNOWN_WIDGET", err.Error())
		case errors.Is(err, dashboard.ErrWidgetNotAllowed):
			httputil.Error(c, http.StatusForbidden, "WIDGET_NOT_ALLOWED", err.Error())
		case errors.Is(err, dashboard.ErrInvalidLayout):
			httputil.Error(c, http.StatusBadRequest, "INVALID_LAYOUT", err.Error())
		default:
			httputil.InternalError(c, "unable to save dashboard layout")
		}
		return
	}

	httputil.JSON(c, http.StatusOK, result)
}

func (h *DashboardHandler) DeleteMe(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httputil.Unauthorized(c, "unauthorized")
		return
	}

	if err := h.service.ResetLayout(c.Request.Context(), claims.UserID); err != nil {
		httputil.InternalError(c, "unable to reset dashboard layout")
		return
	}

	c.Status(http.StatusNoContent)
}
