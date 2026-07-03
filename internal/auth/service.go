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

	if user.DisabledAt != nil {
		return nil, ErrUserDisabled
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

	if user.DisabledAt != nil {
		return nil, ErrUserDisabled
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
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// ResolveClaims reloads the user from the database so disabled accounts and role
// changes take effect without waiting for JWT expiry.
func (s *Service) ResolveClaims(tokenClaims *Claims) (*Claims, error) {
	user, err := s.GetUserByID(tokenClaims.UserID)
	if err != nil {
		return nil, err
	}
	if user.DisabledAt != nil {
		return nil, ErrUserDisabled
	}
	return &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
	}, nil
}

// EmailsByIDs returns a map of user id → email for the given ids (deduplicated).
func (s *Service) EmailsByIDs(ids []uuid.UUID) (map[uuid.UUID]string, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]string{}, nil
	}

	unique := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		unique[id] = struct{}{}
	}
	deduped := make([]uuid.UUID, 0, len(unique))
	for id := range unique {
		deduped = append(deduped, id)
	}

	var rows []struct {
		ID    uuid.UUID
		Email string
	}
	if err := s.db.Table("users").Select("id, email").Where("id IN ?", deduped).Scan(&rows).Error; err != nil {
		return nil, err
	}

	out := make(map[uuid.UUID]string, len(rows))
	for _, row := range rows {
		out[row.ID] = row.Email
	}
	return out, nil
}

func (s *Service) ListUsers() ([]UserResponse, error) {
	var users []User
	if err := s.db.Order("created_at ASC").Find(&users).Error; err != nil {
		return nil, err
	}

	responses := make([]UserResponse, len(users))
	for i := range users {
		responses[i] = ToUserResponse(&users[i])
	}
	return responses, nil
}

type PatchUserInput struct {
	Role     *Role
	Disabled *bool
	Password *string
}

func (s *Service) PatchUser(id uuid.UUID, input PatchUserInput) (*User, error) {
	if input.Role == nil && input.Disabled == nil && input.Password == nil {
		return nil, ErrEmptyPatch
	}

	user, err := s.GetUserByID(id)
	if err != nil {
		return nil, err
	}

	if input.Role != nil {
		if *input.Role != RoleAdmin && *input.Role != RoleUser {
			return nil, ErrInvalidRole
		}
		if user.Role == RoleAdmin && *input.Role == RoleUser {
			if err := s.ensureAnotherActiveAdmin(user.ID); err != nil {
				return nil, err
			}
		}
		user.Role = *input.Role
	}

	if input.Disabled != nil {
		if *input.Disabled {
			if user.Role == RoleAdmin && user.DisabledAt == nil {
				if err := s.ensureAnotherActiveAdmin(user.ID); err != nil {
					return nil, err
				}
			}
			now := time.Now()
			user.DisabledAt = &now
			if err := s.revokeAllRefreshTokens(user.ID); err != nil {
				return nil, err
			}
		} else {
			user.DisabledAt = nil
		}
	}

	if input.Password != nil {
		if len(*input.Password) < 8 {
			return nil, ErrInvalidPassword
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("hash password: %w", err)
		}
		user.PasswordHash = string(hash)
	}

	if err := s.db.Save(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) countActiveAdmins(excludeID uuid.UUID) (int64, error) {
	var count int64
	err := s.db.Model(&User{}).
		Where("role = ? AND disabled_at IS NULL AND id != ?", RoleAdmin, excludeID).
		Count(&count).Error
	return count, err
}

func (s *Service) ensureAnotherActiveAdmin(userID uuid.UUID) error {
	count, err := s.countActiveAdmins(userID)
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrLastAdmin
	}
	return nil
}

func (s *Service) revokeAllRefreshTokens(userID uuid.UUID) error {
	now := time.Now()
	return s.db.Model(&RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
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

// AuthenticateForVolume validates password against the volume owner, then any active admin.
// Used by FTP where the username is the volume UUID.
func (s *Service) AuthenticateForVolume(volumeID uuid.UUID, password string) (*Claims, error) {
	var vol struct {
		OwnerID uuid.UUID
	}
	if err := s.db.Table("volumes").Select("owner_id").Where("id = ?", volumeID).First(&vol).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	owner, err := s.GetUserByID(vol.OwnerID)
	if err != nil {
		return nil, err
	}
	if owner.DisabledAt == nil {
		if err := bcrypt.CompareHashAndPassword([]byte(owner.PasswordHash), []byte(password)); err == nil {
			return &Claims{UserID: owner.ID, Email: owner.Email, Role: owner.Role}, nil
		}
	}

	var admins []User
	if err := s.db.Where("role = ? AND disabled_at IS NULL", RoleAdmin).Find(&admins).Error; err != nil {
		return nil, err
	}
	for i := range admins {
		if err := bcrypt.CompareHashAndPassword([]byte(admins[i].PasswordHash), []byte(password)); err == nil {
			return &Claims{UserID: admins[i].ID, Email: admins[i].Email, Role: admins[i].Role}, nil
		}
	}

	return nil, ErrInvalidCredentials
}

// ValidateCredentials checks email/password without issuing tokens.
func (s *Service) ValidateCredentials(email, password string) (*Claims, error) {
	var user User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if user.DisabledAt != nil {
		return nil, ErrUserDisabled
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return &Claims{UserID: user.ID, Email: user.Email, Role: user.Role}, nil
}
