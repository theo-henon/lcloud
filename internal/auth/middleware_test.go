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

func TestAuthMiddlewareRejectsDisabledUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	service := newTestService(t, db)

	user, err := service.CreateUser("user@example.com", "password123", RoleUser)
	require.NoError(t, err)

	token, err := service.issueAccessToken(user)
	require.NoError(t, err)

	disabled := true
	_, err = service.PatchUser(user.ID, PatchUserInput{Disabled: &disabled})
	require.NoError(t, err)

	router := gin.New()
	router.GET("/protected", AuthMiddleware(service), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddlewareSyncsDemotedAdminRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	service := newTestService(t, db)
	require.NoError(t, service.SeedAdmin("admin@example.com", "adminpass1"))

	users, err := service.ListUsers()
	require.NoError(t, err)
	require.Len(t, users, 1)
	soleAdminID := users[0].ID

	member, err := service.CreateUser("member@example.com", "password123", RoleUser)
	require.NoError(t, err)

	promoted := RoleAdmin
	_, err = service.PatchUser(member.ID, PatchUserInput{Role: &promoted})
	require.NoError(t, err)

	soleAdmin, err := service.GetUserByID(soleAdminID)
	require.NoError(t, err)
	token, err := service.issueAccessToken(soleAdmin)
	require.NoError(t, err)

	demoted := RoleUser
	_, err = service.PatchUser(soleAdminID, PatchUserInput{Role: &demoted})
	require.NoError(t, err)

	router := gin.New()
	router.GET("/admin", AuthMiddleware(service), RequireAdmin(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
}
