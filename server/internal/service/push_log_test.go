package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/internal/model"
)

func TestLogPushSender_Send_PushLog(t *testing.T) {
	sender := NewLogPushSender()
	ctx := context.Background()

	t.Run("send with long token", func(t *testing.T) {
		err := sender.Send(ctx, "abcdefghijklmnopqrstuvwxyz", model.Notification{
			Type:  model.NotifTypeMessage,
			Title: "Test",
			Body:  "Hello",
		})
		require.NoError(t, err)
	})

	t.Run("send with short token", func(t *testing.T) {
		err := sender.Send(ctx, "short", model.Notification{
			Type:  model.NotifTypeMessage,
			Title: "Test",
			Body:  "Hello",
		})
		require.NoError(t, err)
	})
}

func TestLogPushSender_SendMulti_PushLog(t *testing.T) {
	sender := NewLogPushSender()
	ctx := context.Background()

	err := sender.SendMulti(ctx, []string{"token1", "token2"}, model.Notification{
		Type:  model.NotifTypeMessage,
		Title: "Multi",
		Body:  "Broadcast",
	})
	require.NoError(t, err)
}

func TestMin(t *testing.T) {
	assert.Equal(t, 3, min(3, 5))
	assert.Equal(t, 3, min(5, 3))
	assert.Equal(t, 0, min(0, 0))
}
