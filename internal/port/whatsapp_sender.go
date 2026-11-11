package port

import "context"

// WhatsAppSender sends messages through an existing WhatsApp session. It is
// narrower than WhatsAppGateway so use cases can reply without depending on
// session lifecycle operations.
type WhatsAppSender interface {
	SendMessageContext(ctx context.Context, sessionID uint, to string, content string) error
	SendPoll(ctx context.Context, sessionID uint, to string, question string, options []string, selectable int) (string, error)
	SendChatPresence(ctx context.Context, sessionID uint, to string, state string, media string) error
}
