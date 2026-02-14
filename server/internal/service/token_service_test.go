package service_test

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/internal/service"
)

func setupTokenTest(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	return s, client
}

func TestTokenService_Generate(t *testing.T) {
	_, client := setupTokenTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewTokenService(client, service.DefaultTokenConfig("test-secret-key-123"))
	ctx := context.Background()
	userID := uuid.New()

	tokens, err := svc.Generate(ctx, userID)
	require.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	assert.Greater(t, tokens.ExpiresAt, int64(0))
}

func TestTokenService_Validate(t *testing.T) {
	_, client := setupTokenTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewTokenService(client, service.DefaultTokenConfig("test-secret-key-123"))
	ctx := context.Background()
	userID := uuid.New()

	tokens, err := svc.Generate(ctx, userID)
	require.NoError(t, err)

	claims, err := svc.Validate(tokens.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
}

func TestTokenService_Validate_Invalid(t *testing.T) {
	_, client := setupTokenTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewTokenService(client, service.DefaultTokenConfig("test-secret-key-123"))

	_, err := svc.Validate("invalid.token.here")
	assert.Error(t, err)
}

func TestTokenService_Validate_WrongSecret(t *testing.T) {
	_, client := setupTokenTest(t)
	defer func() { _ = client.Close() }()

	svc1 := service.NewTokenService(client, service.DefaultTokenConfig("secret-one"))
	svc2 := service.NewTokenService(client, service.DefaultTokenConfig("secret-two"))
	ctx := context.Background()

	tokens, err := svc1.Generate(ctx, uuid.New())
	require.NoError(t, err)

	_, err = svc2.Validate(tokens.AccessToken)
	assert.Error(t, err)
}

func TestTokenService_Refresh(t *testing.T) {
	_, client := setupTokenTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewTokenService(client, service.DefaultTokenConfig("test-secret-key-123"))
	ctx := context.Background()
	userID := uuid.New()

	original, err := svc.Generate(ctx, userID)
	require.NoError(t, err)

	newTokens, err := svc.Refresh(ctx, original.RefreshToken)
	require.NoError(t, err)
	assert.NotEqual(t, original.AccessToken, newTokens.AccessToken)
	assert.NotEqual(t, original.RefreshToken, newTokens.RefreshToken)

	// Old refresh token should be revoked
	_, err = svc.Refresh(ctx, original.RefreshToken)
	assert.Error(t, err)
}

func TestTokenService_Revoke(t *testing.T) {
	_, client := setupTokenTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewTokenService(client, service.DefaultTokenConfig("test-secret-key-123"))
	ctx := context.Background()

	tokens, err := svc.Generate(ctx, uuid.New())
	require.NoError(t, err)

	err = svc.Revoke(ctx, tokens.AccessToken, tokens.RefreshToken)
	assert.NoError(t, err)

	// Refresh should fail now
	_, err = svc.Refresh(ctx, tokens.RefreshToken)
	assert.Error(t, err)
}

func TestTokenService_Refresh_InvalidToken(t *testing.T) {
	_, client := setupTokenTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewTokenService(client, service.DefaultTokenConfig("test-secret-key-123"))
	ctx := context.Background()

	_, err := svc.Refresh(ctx, "invalid-token")
	assert.Error(t, err)
}

func TestTokenService_Refresh_RevokedToken(t *testing.T) {
	s, client := setupTokenTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewTokenService(client, service.DefaultTokenConfig("test-secret-key-123"))
	ctx := context.Background()
	userID := uuid.New()

	tokens, err := svc.Generate(ctx, userID)
	require.NoError(t, err)

	// Delete the refresh token from Redis to simulate revocation
	claims, err := svc.Validate(tokens.RefreshToken)
	require.NoError(t, err)
	s.Del("refresh:" + claims.ID)

	_, err = svc.Refresh(ctx, tokens.RefreshToken)
	assert.Error(t, err)
}

func TestTokenService_Refresh_MismatchedUser(t *testing.T) {
	s, client := setupTokenTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewTokenService(client, service.DefaultTokenConfig("test-secret-key-123"))
	ctx := context.Background()
	userID := uuid.New()

	tokens, err := svc.Generate(ctx, userID)
	require.NoError(t, err)

	// Tamper with Redis value to a different user
	claims, err := svc.Validate(tokens.RefreshToken)
	require.NoError(t, err)
	s.Set("refresh:"+claims.ID, uuid.New().String())

	_, err = svc.Refresh(ctx, tokens.RefreshToken)
	assert.Error(t, err)
}

func TestTokenService_Revoke_EmptyTokens(t *testing.T) {
	_, client := setupTokenTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewTokenService(client, service.DefaultTokenConfig("test-secret-key-123"))
	ctx := context.Background()

	// Revoke with empty tokens should not error
	err := svc.Revoke(ctx, "", "")
	assert.NoError(t, err)
}

func TestTokenService_Revoke_InvalidTokens(t *testing.T) {
	_, client := setupTokenTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewTokenService(client, service.DefaultTokenConfig("test-secret-key-123"))
	ctx := context.Background()

	// Revoke with invalid tokens should not error (invalid tokens are ignored)
	err := svc.Revoke(ctx, "invalid-access", "invalid-refresh")
	assert.NoError(t, err)
}

func TestTokenService_Generate_RedisError(t *testing.T) {
	s, client := setupTokenTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewTokenService(client, service.DefaultTokenConfig("test-secret-key-123"))

	// Close miniredis to cause Redis error
	s.Close()

	_, err := svc.Generate(context.Background(), uuid.New())
	assert.Error(t, err)
}
