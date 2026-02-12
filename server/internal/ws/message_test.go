package ws_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/internal/ws"
)

func TestWSMessage_MarshalUnmarshal(t *testing.T) {
	msg := ws.WSMessage{
		Type:    ws.WSTypeMessage,
		Payload: json.RawMessage(`{"text":"hello"}`),
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var decoded ws.WSMessage
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, ws.WSTypeMessage, decoded.Type)
	assert.JSONEq(t, `{"text":"hello"}`, string(decoded.Payload))
}

func TestWSMessageTypes(t *testing.T) {
	assert.Equal(t, "message", ws.WSTypeMessage)
	assert.Equal(t, "typing", ws.WSTypeTyping)
	assert.Equal(t, "online_status", ws.WSTypeOnlineStatus)
	assert.Equal(t, "read_receipt", ws.WSTypeReadReceipt)
	assert.Equal(t, "doc_update", ws.WSTypeDocUpdate)
	assert.Equal(t, "doc_lock", ws.WSTypeDocLock)
	assert.Equal(t, "notification", ws.WSTypeNotification)
}
