package auth

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func setupTestRedis(t *testing.T) (*redis.Client, func()) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return rdb, func() {
		rdb.Close()
		mr.Close()
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	token1, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	token2, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	if token1 == "" {
		t.Error("Generated token should not be empty")
	}

	if token1 == token2 {
		t.Error("Generated tokens should be unique")
	}

	// Token should be base64 URL encoded (43 chars for 32 bytes)
	if len(token1) < 40 {
		t.Errorf("Token seems too short: %d chars", len(token1))
	}
}

func TestStoreAndGetRefreshToken(t *testing.T) {
	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	mgr := NewRefreshTokenManager(rdb)
	ctx := context.Background()

	userID := uuid.New()
	email := "test@example.com"
	name := "Test User"

	token, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Store the token
	err = mgr.StoreRefreshToken(ctx, token, userID, email, name)
	if err != nil {
		t.Fatalf("Failed to store token: %v", err)
	}

	// Retrieve the token
	data, err := mgr.GetRefreshToken(ctx, token)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	if data.UserID != userID {
		t.Errorf("Expected user ID %v, got %v", userID, data.UserID)
	}

	if data.Email != email {
		t.Errorf("Expected email %s, got %s", email, data.Email)
	}

	if data.Name != name {
		t.Errorf("Expected name %s, got %s", name, data.Name)
	}

	if data.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestGetInvalidRefreshToken(t *testing.T) {
	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	mgr := NewRefreshTokenManager(rdb)
	ctx := context.Background()

	_, err := mgr.GetRefreshToken(ctx, "nonexistent-token")
	if err != ErrInvalidRefreshToken {
		t.Errorf("Expected ErrInvalidRefreshToken, got %v", err)
	}
}

func TestRevokeRefreshToken(t *testing.T) {
	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	mgr := NewRefreshTokenManager(rdb)
	ctx := context.Background()

	userID := uuid.New()
	token, _ := GenerateRefreshToken()

	// Store the token
	err := mgr.StoreRefreshToken(ctx, token, userID, "test@example.com", "Test")
	if err != nil {
		t.Fatalf("Failed to store token: %v", err)
	}

	// Verify it exists
	_, err = mgr.GetRefreshToken(ctx, token)
	if err != nil {
		t.Fatalf("Token should exist: %v", err)
	}

	// Revoke the token
	err = mgr.RevokeRefreshToken(ctx, token)
	if err != nil {
		t.Fatalf("Failed to revoke token: %v", err)
	}

	// Verify it's gone
	_, err = mgr.GetRefreshToken(ctx, token)
	if err != ErrInvalidRefreshToken {
		t.Error("Token should be invalid after revocation")
	}
}

func TestRevokeAllUserTokens(t *testing.T) {
	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	mgr := NewRefreshTokenManager(rdb)
	ctx := context.Background()

	userID := uuid.New()
	email := "test@example.com"

	// Create multiple tokens for the same user
	tokens := make([]string, 3)
	for i := 0; i < 3; i++ {
		token, _ := GenerateRefreshToken()
		tokens[i] = token
		err := mgr.StoreRefreshToken(ctx, token, userID, email, "Test")
		if err != nil {
			t.Fatalf("Failed to store token %d: %v", i, err)
		}
	}

	// Verify all tokens exist
	for i, token := range tokens {
		_, err := mgr.GetRefreshToken(ctx, token)
		if err != nil {
			t.Fatalf("Token %d should exist: %v", i, err)
		}
	}

	// Revoke all tokens
	err := mgr.RevokeAllUserTokens(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to revoke all tokens: %v", err)
	}

	// Verify all tokens are gone
	for i, token := range tokens {
		_, err := mgr.GetRefreshToken(ctx, token)
		if err != ErrInvalidRefreshToken {
			t.Errorf("Token %d should be invalid after logout-all", i)
		}
	}
}

func TestRotateRefreshToken(t *testing.T) {
	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	mgr := NewRefreshTokenManager(rdb)
	ctx := context.Background()

	userID := uuid.New()
	email := "test@example.com"
	name := "Test User"

	oldToken, _ := GenerateRefreshToken()

	// Store the original token
	err := mgr.StoreRefreshToken(ctx, oldToken, userID, email, name)
	if err != nil {
		t.Fatalf("Failed to store token: %v", err)
	}

	// Rotate the token
	newToken, data, err := mgr.RotateRefreshToken(ctx, oldToken)
	if err != nil {
		t.Fatalf("Failed to rotate token: %v", err)
	}

	if newToken == oldToken {
		t.Error("New token should be different from old token")
	}

	if data.UserID != userID {
		t.Errorf("Expected user ID %v, got %v", userID, data.UserID)
	}

	// Old token should be invalid
	_, err = mgr.GetRefreshToken(ctx, oldToken)
	if err != ErrInvalidRefreshToken {
		t.Error("Old token should be invalid after rotation")
	}

	// New token should be valid
	newData, err := mgr.GetRefreshToken(ctx, newToken)
	if err != nil {
		t.Fatalf("New token should be valid: %v", err)
	}

	if newData.UserID != userID {
		t.Errorf("Expected user ID %v, got %v", userID, newData.UserID)
	}
}

func TestRotateInvalidToken(t *testing.T) {
	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	mgr := NewRefreshTokenManager(rdb)
	ctx := context.Background()

	_, _, err := mgr.RotateRefreshToken(ctx, "nonexistent-token")
	if err != ErrInvalidRefreshToken {
		t.Errorf("Expected ErrInvalidRefreshToken, got %v", err)
	}
}

func TestRefreshTokenExpiry(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer rdb.Close()

	mgr := NewRefreshTokenManager(rdb)
	ctx := context.Background()

	userID := uuid.New()
	token, _ := GenerateRefreshToken()

	err = mgr.StoreRefreshToken(ctx, token, userID, "test@example.com", "Test")
	if err != nil {
		t.Fatalf("Failed to store token: %v", err)
	}

	// Check TTL is set
	ttl := mr.TTL("refresh_token:" + token)
	if ttl <= 0 {
		t.Error("Token should have TTL set")
	}

	// TTL should be close to 30 days
	expectedTTL := 30 * 24 * time.Hour
	if ttl < expectedTTL-time.Minute || ttl > expectedTTL+time.Minute {
		t.Errorf("TTL should be around 30 days, got %v", ttl)
	}
}
