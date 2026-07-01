package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	service := newTestService(t, db)

	user, err := service.CreateUser("user@example.com", "password123", RoleUser)
	require.NoError(t, err)

	router := gin.New()
	router.GET("/protected", AuthMiddleware(service), func(c *gin.Context) {
		claims, ok := ClaimsFromContext(c)
		require.True(t, ok)
		c.JSON(http.StatusOK, gin.H{"email": claims.Email})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)

	token, err := service.issueAccessToken(user)
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/admin", func(c *gin.Context) {
		c.Set(claimsContextKey, &Claims{Role: RoleUser})
		c.Next()
	}, RequireAdmin(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestValidateAccessTokenExpired(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(t, db)

	claims := jwtClaims{
		Email: "user@example.com",
		Role:  RoleUser,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   uuid.New().String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(service.jwtSecret)
	require.NoError(t, err)

	_, err = service.ValidateAccessToken(signed)
	require.ErrorIs(t, err, ErrExpiredToken)
}

func TestLogoutInvalidToken(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(t, db)

	err := service.Logout("invalid")
	require.ErrorIs(t, err, ErrInvalidToken)
}
