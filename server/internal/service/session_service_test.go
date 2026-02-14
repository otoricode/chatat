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

func setupSessionTest(t *testing.T) (*miniredis.Miniredis, *redis.Client, service.TokenService) {
	t.Helper()
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	tokenSvc := service.NewTokenService(client, service.DefaultTokenConfig("test-session-secret"))
	return s, client, tokenSvc
}

func TestSessionService_Register(t *testing.T) {
	_, client, tokenSvc := setupSessionTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewSessionService(client, tokenSvc, 0)
	ctx := context.Background()
	userID := uuid.New()

	err := svc.Register(ctx, userID, "device-1", "refresh-token-1")
	require.NoError(t, err)

	err = svc.Validate(ctx, userID, "device-1")
	assert.NoError(t, err)
}

func TestSessionService_Validate_WrongDevice(t *testing.T) {
	_, client, tokenSvc := setupSessionTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewSessionService(client, tokenSvc, 0)
	ctx := context.Background()
	userID := uuid.New()

	err := svc.Register(ctx, userID, "device-1", "refresh-token-1")
	require.NoError(t, err)

	err = svc.Validate(ctx, userID, "device-2")
	assert.Error(t, err)
}

func TestSessionService_OneDevicePerUser(t *testing.T) {
	_, client, tokenSvc := setupSessionTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewSessionService(client, tokenSvc, 0)
	ctx := context.Background()
	userID := uuid.New()

	// Register device 1
	err := svc.Register(ctx, userID, "device-1", "refresh-token-1")
	require.NoError(t, err)

	// Register device 2 (should revoke device 1)
	err = svc.Register(ctx, userID, "device-2", "refresh-token-2")
	require.NoError(t, err)

	// Device 1 should no longer be valid
	err = svc.Validate(ctx, userID, "device-1")
	assert.Error(t, err)

	// Device 2 should be valid
	err = svc.Validate(ctx, userID, "device-2")
	assert.NoError(t, err)
}

func TestSessionService_Invalidate(t *testing.T) {
	_, client, tokenSvc := setupSessionTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewSessionService(client, tokenSvc, 0)
	ctx := context.Background()
	userID := uuid.New()

	err := svc.Register(ctx, userID, "device-1", "refresh-token-1")
	require.NoError(t, err)

	err = svc.Invalidate(ctx, userID)
	require.NoError(t, err)

	err = svc.Validate(ctx, userID, "device-1")
	assert.Error(t, err)
}

func TestSessionService_Validate_NoSession(t *testing.T) {
	_, client, tokenSvc := setupSessionTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewSessionService(client, tokenSvc, 0)
	ctx := context.Background()

	err := svc.Validate(ctx, uuid.New(), "device-1")
	assert.Error(t, err)
}

func TestSessionService_Register_RedisError(t *testing.T) {
	s, client, tokenSvc := setupSessionTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewSessionService(client, tokenSvc, 0)

	// Close miniredis to cause errors
	s.Close()

	err := svc.Register(context.Background(), uuid.New(), "device-1", "refresh-1")
	assert.Error(t, err)
}

func TestSessionService_Validate_RedisError(t *testing.T) {
	s, client, tokenSvc := setupSessionTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewSessionService(client, tokenSvc, 0)
	userID := uuid.New()

	// Register first
	err := svc.Register(context.Background(), userID, "device-1", "refresh-1")
	require.NoError(t, err)

	// Close miniredis to cause error on validate
	s.Close()

	err = svc.Validate(context.Background(), userID, "device-1")
	assert.Error(t, err)
}

func TestSessionService_Register_SameDevice(t *testing.T) {
	_, client, tokenSvc := setupSessionTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewSessionService(client, tokenSvc, 0)
	ctx := context.Background()
	userID := uuid.New()

	// Register same device twice - should not revoke
	err := svc.Register(ctx, userID, "device-1", "refresh-1")
	require.NoError(t, err)

	err = svc.Register(ctx, userID, "device-1", "refresh-2")
	require.NoError(t, err)

	// Should still be valid
	err = svc.Validate(ctx, userID, "device-1")
	assert.NoError(t, err)
}
