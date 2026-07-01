package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrExpiredToken       = errors.New("expired token")
	ErrRevokedToken       = errors.New("revoked token")
	ErrEmailTaken         = errors.New("email already taken")
	ErrInvalidRole        = errors.New("invalid role")
)

type jwtClaims struct {
	Email string `json:"email"`
	Role  Role   `json:"role"`
	jwt.RegisteredClaims
}

func (s *Service) SeedAdmin(email, password string) error {
	var count int64
	if err := s.db.Model(&User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err := s.createUser(email, password, RoleAdmin)
	return err
}

func (s *Service) CreateUser(email, password string, role Role) (*User, error) {
	if role == "" {
		role = RoleUser
	}
	if role != RoleAdmin && role != RoleUser {
		return nil, ErrInvalidRole
	}
	return s.createUser(email, password, role)
}

func (s *Service) createUser(email, password string, role Role) (*User, error) {
	var existing User
	if err := s.db.Where("email = ?", email).First(&existing).Error; err == nil {
		return nil, ErrEmailTaken
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) Login(email, password string) (*LoginResult, error) {
	var user User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, err := s.issueAccessToken(&user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.issueRefreshToken(&user)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         ToUserResponse(&user),
	}, nil
}

func (s *Service) Refresh(refreshToken string) (*RefreshResult, error) {
	record, err := s.findRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	var user User
	if err := s.db.First(&user, "id = ?", record.UserID).Error; err != nil {
		return nil, ErrInvalidToken
	}

	now := time.Now()
	record.RevokedAt = &now
	if err := s.db.Save(record).Error; err != nil {
		return nil, err
	}

	accessToken, err := s.issueAccessToken(&user)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.issueRefreshToken(&user)
	if err != nil {
		return nil, err
	}

	return &RefreshResult{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *Service) Logout(refreshToken string) error {
	record, err := s.findRefreshToken(refreshToken)
	if err != nil {
		return err
	}

	now := time.Now()
	record.RevokedAt = &now
	return s.db.Save(record).Error
}

func (s *Service) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, ErrInvalidToken
	}

	return &Claims{
		UserID: userID,
		Email:  claims.Email,
		Role:   claims.Role,
	}, nil
}

func (s *Service) GetUserByID(id uuid.UUID) (*User, error) {
	var user User
	if err := s.db.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}
	return &user, nil
}

func (s *Service) issueAccessToken(user *User) (string, error) {
	now := time.Now()
	claims := jwtClaims{
		Email: user.Email,
		Role:  user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtExpiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *Service) issueRefreshToken(user *User) (string, error) {
	raw, err := generateToken()
	if err != nil {
		return "", err
	}

	record := &RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: hashToken(raw),
		ExpiresAt: time.Now().Add(time.Duration(s.refreshTokenExpiryDays) * 24 * time.Hour),
	}

	if err := s.db.Create(record).Error; err != nil {
		return "", err
	}

	return raw, nil
}

func (s *Service) findRefreshToken(raw string) (*RefreshToken, error) {
	var record RefreshToken
	if err := s.db.Where("token_hash = ?", hashToken(raw)).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}

	if record.RevokedAt != nil {
		return nil, ErrRevokedToken
	}
	if time.Now().After(record.ExpiresAt) {
		return nil, ErrExpiredToken
	}

	return &record, nil
}

func generateToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
