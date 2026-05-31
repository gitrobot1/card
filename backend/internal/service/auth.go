package service

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	appconfig "github.com/time/card/backend/internal/config"
	"github.com/time/card/backend/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

var (
	ErrInvalidUsername = errors.New("invalid username")
	usernamePattern    = regexp.MustCompile(`^[\p{Han}a-zA-Z0-9_]+$`)
)

type AuthService struct {
	db     *gorm.DB
	secret []byte
	ttl    time.Duration
}

type AuthResult struct {
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	User      model.User `json:"user"`
	IsNewUser bool       `json:"is_new_user"`
}

type tokenClaims struct {
	UserID   uint64 `json:"uid"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func NewAuthService(db *gorm.DB, cfg *appconfig.AuthConfig) *AuthService {
	return &AuthService{
		db:     db,
		secret: []byte(cfg.JWTSecret),
		ttl:    cfg.TokenDuration(),
	}
}

func (s *AuthService) LoginByUsername(username string) (*AuthResult, error) {
	normalized, err := normalizeUsername(username)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var user model.User
	err = s.db.Where("username = ?", normalized).First(&user).Error
	isNewUser := errors.Is(err, gorm.ErrRecordNotFound)
	if err != nil && !isNewUser {
		return nil, err
	}

	if isNewUser {
		user = model.User{
			Username:  normalized,
			Nickname:  normalized,
			LastLogin: now,
		}
		if err := s.db.Create(&user).Error; err != nil {
			return nil, err
		}
	} else {
		user.LastLogin = now
		if err := s.db.Model(&user).Update("last_login", now).Error; err != nil {
			return nil, err
		}
	}

	token, expiresAt, err := s.issueToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
		IsNewUser: isNewUser,
	}, nil
}

func (s *AuthService) ParseToken(token string) (*tokenClaims, error) {
	parsed, err := jwt.ParseWithClaims(token, &tokenClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := parsed.Claims.(*tokenClaims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func (s *AuthService) GetUserByID(id uint64) (*model.User, error) {
	var user model.User
	if err := s.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *AuthService) issueToken(user model.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.ttl)
	claims := tokenClaims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", user.ID),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}

func normalizeUsername(raw string) (string, error) {
	username := strings.TrimSpace(raw)
	length := utf8.RuneCountInString(username)
	if length < 2 || length > 16 {
		return "", ErrInvalidUsername
	}
	if !usernamePattern.MatchString(username) {
		return "", ErrInvalidUsername
	}
	return username, nil
}
