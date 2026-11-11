package whatsapp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type receivedMessage struct {
	sessionID uint
	messageID string
	chatJID   string
	sender    string
	content   string
}

func TestMakeMessageCallbackForwardsMessageIdentity(t *testing.T) {
	handler := NewEventHandler(nil, nil)
	got := make(chan receivedMessage, 1)

	callback := handler.MakeMessageCallback(func(
		_ context.Context,
		sessionID uint,
		messageID string,
		chatJID string,
		customerNumber string,
		content string,
	) error {
		got <- receivedMessage{
			sessionID: sessionID,
			messageID: messageID,
			chatJID:   chatJID,
			sender:    customerNumber,
			content:   content,
		}
		return nil
	})

	callback(7, "msg-1", "628@s.whatsapp.net", "628", "halo")

	select {
	case message := <-got:
		assert.Equal(t, receivedMessage{
			sessionID: 7,
			messageID: "msg-1",
			chatJID:   "628@s.whatsapp.net",
			sender:    "628",
			content:   "halo",
		}, message)
	case <-time.After(time.Second):
		require.FailNow(t, "message callback did not run")
	}
}
