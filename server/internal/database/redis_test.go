package database_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/internal/database"
)

func getRedisURL() string {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		return "redis://localhost:6380"
	}
	return url
}

func TestNewRedisClient_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	client, err := database.NewRedisClient(getRedisURL())
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	assert.NotNil(t, client)
}

func TestNewRedisClient_InvalidURL(t *testing.T) {
	_, err := database.NewRedisClient("invalid://url")
	assert.Error(t, err)
}

func TestRedisClient_SetGetWithTTL(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	client, err := database.NewRedisClient(getRedisURL())
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	err = client.Set(ctx, "test:key", "hello", 2*time.Second).Err()
	require.NoError(t, err)

	val, err := client.Get(ctx, "test:key").Result()
	require.NoError(t, err)
	assert.Equal(t, "hello", val)

	err = client.Del(ctx, "test:key").Err()
	require.NoError(t, err)

	_, err = client.Get(ctx, "test:key").Result()
	assert.Error(t, err)
}

func TestRedisClient_PubSub(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	client, err := database.NewRedisClient(getRedisURL())
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	sub := client.Subscribe(ctx, "test:channel")
	defer func() { _ = sub.Close() }()

	_, err = sub.Receive(ctx)
	require.NoError(t, err)

	err = client.Publish(ctx, "test:channel", "test-message").Err()
	require.NoError(t, err)

	msg, err := sub.ReceiveMessage(ctx)
	require.NoError(t, err)
	assert.Equal(t, "test-message", msg.Payload)
}
