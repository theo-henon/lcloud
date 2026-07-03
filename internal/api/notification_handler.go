package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/theo-henon/lcloud/internal/auth"
	"github.com/theo-henon/lcloud/internal/notification"
	"github.com/theo-henon/lcloud/pkg/httputil"
)

type NotificationHandler struct {
	notifications *notification.Service
}

func NewNotificationHandler(service *notification.Service) *NotificationHandler {
	return &NotificationHandler{notifications: service}
}

func (h *NotificationHandler) List(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httputil.Unauthorized(c, "unauthorized")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	items, total, err := h.notifications.List(c.Request.Context(), claims.UserID, limit, offset)
	if err != nil {
		httputil.InternalError(c, "unable to list notifications")
		return
	}

	httputil.JSON(c, http.StatusOK, gin.H{
		"notifications": items,
		"total":         total,
	})
}

func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httputil.Unauthorized(c, "unauthorized")
		return
	}

	count, err := h.notifications.UnreadCount(c.Request.Context(), claims.UserID)
	if err != nil {
		httputil.InternalError(c, "unable to count notifications")
		return
	}

	httputil.JSON(c, http.StatusOK, gin.H{"count": count})
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httputil.Unauthorized(c, "unauthorized")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.BadRequest(c, "invalid notification id")
		return
	}

	item, err := h.notifications.MarkRead(c.Request.Context(), claims.UserID, id)
	if err != nil {
		if errors.Is(err, notification.ErrNotFound) {
			httputil.Error(c, http.StatusNotFound, "NOT_FOUND", "notification not found")
			return
		}
		httputil.InternalError(c, "unable to mark notification read")
		return
	}

	httputil.JSON(c, http.StatusOK, item)
}

func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httputil.Unauthorized(c, "unauthorized")
		return
	}

	updated, err := h.notifications.MarkAllRead(c.Request.Context(), claims.UserID)
	if err != nil {
		httputil.InternalError(c, "unable to mark notifications read")
		return
	}

	httputil.JSON(c, http.StatusOK, gin.H{"updated": updated})
}

func (h *NotificationHandler) Dismiss(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httputil.Unauthorized(c, "unauthorized")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.BadRequest(c, "invalid notification id")
		return
	}

	if err := h.notifications.Dismiss(c.Request.Context(), claims.UserID, id); err != nil {
		if errors.Is(err, notification.ErrNotFound) {
			httputil.Error(c, http.StatusNotFound, "NOT_FOUND", "notification not found")
			return
		}
		httputil.InternalError(c, "unable to dismiss notification")
		return
	}

	c.Status(http.StatusNoContent)
}
