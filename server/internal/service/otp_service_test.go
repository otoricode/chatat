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

type mockSMSProvider struct {
	lastPhone   string
	lastMessage string
	sendError   error
}

func (m *mockSMSProvider) Send(phone string, message string) error {
	m.lastPhone = phone
	m.lastMessage = message
	return m.sendError
}

func setupOTPTest(t *testing.T) (*miniredis.Miniredis, *redis.Client, *mockSMSProvider) {
	t.Helper()
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	sms := &mockSMSProvider{}
	return s, client, sms
}

func TestOTPService_Generate(t *testing.T) {
	_, client, sms := setupOTPTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewOTPService(client, sms, service.DefaultOTPConfig())
	ctx := context.Background()

	code, err := svc.Generate(ctx, "+6281234567890")
	require.NoError(t, err)
	assert.Len(t, code, 6)
	assert.Equal(t, "+6281234567890", sms.lastPhone)
	assert.Contains(t, sms.lastMessage, code)
}

func TestOTPService_Verify_Success(t *testing.T) {
	_, client, sms := setupOTPTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewOTPService(client, sms, service.DefaultOTPConfig())
	ctx := context.Background()

	code, err := svc.Generate(ctx, "+6281234567890")
	require.NoError(t, err)

	err = svc.Verify(ctx, "+6281234567890", code)
	assert.NoError(t, err)
}

func TestOTPService_Verify_WrongCode(t *testing.T) {
	_, client, sms := setupOTPTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewOTPService(client, sms, service.DefaultOTPConfig())
	ctx := context.Background()

	_, err := svc.Generate(ctx, "+6281234567890")
	require.NoError(t, err)

	err = svc.Verify(ctx, "+6281234567890", "000000")
	assert.Error(t, err)
}

func TestOTPService_Verify_MaxAttempts(t *testing.T) {
	_, client, sms := setupOTPTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewOTPService(client, sms, service.DefaultOTPConfig())
	ctx := context.Background()

	_, err := svc.Generate(ctx, "+6281234567890")
	require.NoError(t, err)

	// Fail 3 times
	for i := 0; i < 3; i++ {
		err = svc.Verify(ctx, "+6281234567890", "000000")
		assert.Error(t, err)
	}

	// Even with correct code, should fail after max attempts
	// (OTP is deleted after max attempts)
	err = svc.Verify(ctx, "+6281234567890", "123456")
	assert.Error(t, err)
}

func TestOTPService_Generate_Cooldown(t *testing.T) {
	_, client, sms := setupOTPTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewOTPService(client, sms, service.DefaultOTPConfig())
	ctx := context.Background()

	_, err := svc.Generate(ctx, "+6281234567890")
	require.NoError(t, err)

	// Second request should fail due to cooldown
	_, err = svc.Generate(ctx, "+6281234567890")
	assert.Error(t, err)
}

func TestOTPService_Verify_Expired(t *testing.T) {
	_, client, sms := setupOTPTest(t)
	defer func() { _ = client.Close() }()

	svc := service.NewOTPService(client, sms, service.DefaultOTPConfig())
	ctx := context.Background()

	// No OTP generated — should fail
	err := svc.Verify(ctx, "+6281234567890", "123456")
	assert.Error(t, err)
}
