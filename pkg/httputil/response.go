package httputil

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type errorBody struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

func JSON(c *gin.Context, status int, payload any) {
	c.JSON(status, payload)
}

func Error(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, errorBody{
		Error: message,
		Code:  code,
	})
}

func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, "FORBIDDEN", message)
}

func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, "CONFLICT", message)
}

func Unprocessable(c *gin.Context, code, message string) {
	Error(c, http.StatusUnprocessableEntity, code, message)
}

func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}
