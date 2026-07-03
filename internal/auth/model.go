package auth

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type User struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Email        string     `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string     `gorm:"not null" json:"-"`
	Role         Role       `gorm:"type:varchar(16);not null;default:user" json:"role"`
	DisabledAt   *time.Time `gorm:"index" json:"disabled_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

type RefreshToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	TokenHash string     `gorm:"not null;index"`
	ExpiresAt time.Time  `gorm:"not null;index"`
	RevokedAt *time.Time `gorm:"index"`
	CreatedAt time.Time
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

type UserResponse struct {
	ID         uuid.UUID  `json:"id"`
	Email      string     `json:"email"`
	Role       Role       `json:"role"`
	CreatedAt  time.Time  `json:"created_at"`
	DisabledAt *time.Time `json:"disabled_at"`
}

func ToUserResponse(user *User) UserResponse {
	return UserResponse{
		ID:         user.ID,
		Email:      user.Email,
		Role:       user.Role,
		CreatedAt:  user.CreatedAt,
		DisabledAt: user.DisabledAt,
	}
}

type Claims struct {
	UserID uuid.UUID
	Email  string
	Role   Role
}

type LoginResult struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

type RefreshResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Service struct {
	db                     *gorm.DB
	jwtSecret              []byte
	jwtExpiry              time.Duration
	refreshTokenExpiryDays int
}

func NewService(db *gorm.DB, jwtSecret string, jwtExpiryHours, refreshTokenExpiryDays int) *Service {
	return &Service{
		db:                     db,
		jwtSecret:              []byte(jwtSecret),
		jwtExpiry:              time.Duration(jwtExpiryHours) * time.Hour,
		refreshTokenExpiryDays: refreshTokenExpiryDays,
	}
}
