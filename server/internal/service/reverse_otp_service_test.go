package service_test

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/internal/service"
)

type mockWAProvider struct {
	businessNumber string
}

func (m *mockWAProvider) GetBusinessNumber() string {
	return m.businessNumber
}

func setupReverseOTPTest(t *testing.T) (*miniredis.Miniredis, *redis.Client, *mockWAProvider) {
	t.Helper()
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	wa := &mockWAProvider{businessNumber: "+628001234567"}
	return s, client, wa
}

func TestReverseOTP_InitSession(t *testing.T) {
	_, client, wa := setupReverseOTPTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewReverseOTPService(client, wa, 0)
	ctx := context.Background()

	session, err := svc.InitSession(ctx, "+6281234567890")
	require.NoError(t, err)
	assert.NotEmpty(t, session.SessionID)
	assert.Equal(t, "+628001234567", session.TargetWANumber)
	assert.Len(t, session.UniqueCode, 6)
}

func TestReverseOTP_CheckVerification_Pending(t *testing.T) {
	_, client, wa := setupReverseOTPTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewReverseOTPService(client, wa, 0)
	ctx := context.Background()

	session, err := svc.InitSession(ctx, "+6281234567890")
	require.NoError(t, err)

	result, err := svc.CheckVerification(ctx, session.SessionID)
	require.NoError(t, err)
	assert.Equal(t, "pending", result.Status)
}

func TestReverseOTP_HandleIncomingMessage_Verify(t *testing.T) {
	_, client, wa := setupReverseOTPTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewReverseOTPService(client, wa, 0)
	ctx := context.Background()

	session, err := svc.InitSession(ctx, "+6281234567890")
	require.NoError(t, err)

	// Simulate incoming WA message with correct code
	err = svc.HandleIncomingMessage(ctx, "+6281234567890", session.UniqueCode)
	require.NoError(t, err)

	// Check that it's verified
	result, err := svc.CheckVerification(ctx, session.SessionID)
	require.NoError(t, err)
	assert.Equal(t, "verified", result.Status)
	assert.Equal(t, "+6281234567890", result.Phone)
}

func TestReverseOTP_CheckVerification_Expired(t *testing.T) {
	_, client, wa := setupReverseOTPTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewReverseOTPService(client, wa, 0)
	ctx := context.Background()

	// Non-existent session
	result, err := svc.CheckVerification(ctx, "nonexistent-session-id")
	require.NoError(t, err)
	assert.Equal(t, "expired", result.Status)
}

func TestReverseOTP_HandleIncomingMessage_WrongCode(t *testing.T) {
	_, client, wa := setupReverseOTPTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewReverseOTPService(client, wa, 0)
	ctx := context.Background()

	session, err := svc.InitSession(ctx, "+6281234567890")
	require.NoError(t, err)

	// Wrong code
	err = svc.HandleIncomingMessage(ctx, "+6281234567890", "WRONG1")
	require.NoError(t, err)

	// Should still be pending
	result, err := svc.CheckVerification(ctx, session.SessionID)
	require.NoError(t, err)
	assert.Equal(t, "pending", result.Status)
}

func TestReverseOTP_InitSession_Cooldown(t *testing.T) {
	_, client, wa := setupReverseOTPTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewReverseOTPService(client, wa, 0)
	ctx := context.Background()

	_, err := svc.InitSession(ctx, "+6281234567890")
	require.NoError(t, err)

	// Second request should fail due to cooldown
	_, err = svc.InitSession(ctx, "+6281234567890")
	assert.Error(t, err)
}
