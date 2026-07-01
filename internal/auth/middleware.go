package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const claimsContextKey = "authClaims"

func AuthMiddleware(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || len(header) < 8 || header[:7] != "Bearer " {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing or invalid authorization header",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		token := header[7:]
		claims, err := service.ValidateAccessToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired access token",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		c.Set(claimsContextKey, claims)
		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := ClaimsFromContext(c)
		if !ok || claims.Role != RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "admin access required",
				"code":  "FORBIDDEN",
			})
			return
		}
		c.Next()
	}
}

func ClaimsFromContext(c *gin.Context) (*Claims, bool) {
	value, ok := c.Get(claimsContextKey)
	if !ok {
		return nil, false
	}
	claims, ok := value.(*Claims)
	return claims, ok
}
