package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

type TokenManager struct {
	secretKey            []byte
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
	redisClient          *redis.Client
}

// Constructor
func NewTokenManager(secret string, accessDuration, refreshDuration time.Duration, redisClient *redis.Client) *TokenManager {
	return &TokenManager{
		secretKey:            []byte(secret),
		accessTokenDuration:  accessDuration,
		refreshTokenDuration: refreshDuration,
		redisClient:          redisClient,
	}
}

// Token types
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// Claims
type Claims struct {
	UserID    int32  `json:"user_id"`
	Email     string `json:"email"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// Generate Access token
func (m *TokenManager) GenerateToken(userID int32, email string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		Email:     email,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			// Use config from struct
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

// GenerateRefreshToken
func (m *TokenManager) GenerateRefreshToken(userID int32, email string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		Email:     email,
		TokenType: TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

// Validate token (for access tokens only)
func (m *TokenManager) ValidateToken(ctx context.Context, tokenString string) (*Claims, error) {
	// 1. Blacklist checking
	key := "blacklist:" + tokenString
	_, err := m.redisClient.Get(ctx, key).Result()
	if err == nil {
		// Key found = revoked token
		return nil, errors.New("token has been revoked")
	}
	if err != redis.Nil {
		return nil, fmt.Errorf("redis error: %w", err)
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Verify this is an access token
		if claims.TokenType != TokenTypeAccess {
			return nil, errors.New("invalid token type: expected access token")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateRefreshToken validates refresh tokens specifically
func (m *TokenManager) ValidateRefreshToken(ctx context.Context, tokenString string) (*Claims, error) {
	// 1. Blacklist checking
	key := "blacklist:" + tokenString
	_, err := m.redisClient.Get(ctx, key).Result()
	if err == nil {
		return nil, errors.New("token has been revoked")
	}
	if err != redis.Nil {
		return nil, fmt.Errorf("redis error: %w", err)
	}

	// 2. Parse token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Verify this is a refresh token
		if claims.TokenType != TokenTypeRefresh {
			return nil, errors.New("invalid token type: expected refresh token")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// InvalidateToken to blacklist
func (m *TokenManager) InvalidateToken(ctx context.Context, tokenString string) error {
	// 1. Parse token to get exp time without sign
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return errors.New("invalid token claims")
	}

	// 2.
	timeRemaining := time.Until(claims.ExpiresAt.Time)
	if timeRemaining <= 0 {
		return nil
	}

	// 3. Save to redis
	// Key: "blacklist:<token>"
	// value: "revoked"
	// TTL: timeRemaining
	key := "blacklist:" + tokenString
	err = m.redisClient.Set(ctx, key, "revoked", timeRemaining).Err()
	if err != nil {
		return err
	}
	return nil
}
