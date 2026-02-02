package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	RefreshTokenExpiry = 30 * 24 * time.Hour // 30 days
	refreshTokenPrefix = "refresh_token:"
	userTokensPrefix   = "user_tokens:"
)

var (
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrExpiredRefreshToken = errors.New("refresh token has expired")
)

// RefreshTokenData stores the data associated with a refresh token
type RefreshTokenData struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// RefreshTokenManager handles refresh token operations
type RefreshTokenManager struct {
	rdb *redis.Client
}

// NewRefreshTokenManager creates a new refresh token manager
func NewRefreshTokenManager(rdb *redis.Client) *RefreshTokenManager {
	return &RefreshTokenManager{rdb: rdb}
}

// GenerateRefreshToken generates a cryptographically secure refresh token
func GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// StoreRefreshToken stores a refresh token in Valkey
func (m *RefreshTokenManager) StoreRefreshToken(ctx context.Context, token string, userID uuid.UUID, email, name string) error {
	data := RefreshTokenData{
		UserID:    userID,
		Email:     email,
		Name:      name,
		CreatedAt: time.Now(),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Store the token with TTL
	tokenKey := refreshTokenPrefix + token
	if err := m.rdb.Set(ctx, tokenKey, jsonData, RefreshTokenExpiry).Err(); err != nil {
		return err
	}

	// Add token to user's token set for logout-all functionality
	userTokensKey := userTokensPrefix + userID.String()
	if err := m.rdb.SAdd(ctx, userTokensKey, token).Err(); err != nil {
		return err
	}

	// Set expiry on user's token set (should be at least as long as max token expiry)
	// We refresh this on every new token to ensure it doesn't expire while tokens are still valid
	if err := m.rdb.Expire(ctx, userTokensKey, RefreshTokenExpiry+24*time.Hour).Err(); err != nil {
		return err
	}

	return nil
}

// GetRefreshToken retrieves and validates a refresh token
func (m *RefreshTokenManager) GetRefreshToken(ctx context.Context, token string) (*RefreshTokenData, error) {
	tokenKey := refreshTokenPrefix + token

	jsonData, err := m.rdb.Get(ctx, tokenKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, err
	}

	var data RefreshTokenData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

// RevokeRefreshToken revokes a single refresh token
func (m *RefreshTokenManager) RevokeRefreshToken(ctx context.Context, token string) error {
	tokenKey := refreshTokenPrefix + token

	// Get the token data first to find the user
	jsonData, err := m.rdb.Get(ctx, tokenKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// Token doesn't exist, consider it revoked
			return nil
		}
		return err
	}

	var data RefreshTokenData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return err
	}

	// Delete the token
	if err := m.rdb.Del(ctx, tokenKey).Err(); err != nil {
		return err
	}

	// Remove from user's token set
	userTokensKey := userTokensPrefix + data.UserID.String()
	if err := m.rdb.SRem(ctx, userTokensKey, token).Err(); err != nil {
		return err
	}

	return nil
}

// RevokeAllUserTokens revokes all refresh tokens for a user (logout-all)
func (m *RefreshTokenManager) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	userTokensKey := userTokensPrefix + userID.String()

	// Get all tokens for this user
	tokens, err := m.rdb.SMembers(ctx, userTokensKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		return err
	}

	// Delete each token
	for _, token := range tokens {
		tokenKey := refreshTokenPrefix + token
		if err := m.rdb.Del(ctx, tokenKey).Err(); err != nil {
			return err
		}
	}

	// Delete the user's token set
	if err := m.rdb.Del(ctx, userTokensKey).Err(); err != nil {
		return err
	}

	return nil
}

// RotateRefreshToken invalidates the old token and creates a new one
func (m *RefreshTokenManager) RotateRefreshToken(ctx context.Context, oldToken string) (string, *RefreshTokenData, error) {
	// Get the old token data
	data, err := m.GetRefreshToken(ctx, oldToken)
	if err != nil {
		return "", nil, err
	}

	// Revoke the old token
	if err := m.RevokeRefreshToken(ctx, oldToken); err != nil {
		return "", nil, err
	}

	// Generate a new token
	newToken, err := GenerateRefreshToken()
	if err != nil {
		return "", nil, err
	}

	// Store the new token
	if err := m.StoreRefreshToken(ctx, newToken, data.UserID, data.Email, data.Name); err != nil {
		return "", nil, err
	}

	return newToken, data, nil
}
