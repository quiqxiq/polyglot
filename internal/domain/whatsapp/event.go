package whatsapp

// InboundMessage represents the content and type of an incoming message.
type InboundMessage struct {
	Type     string `json:"type"` // "text", "image", "document", etc.
	Text     string `json:"text,omitempty"`
	MediaURL string `json:"media_url,omitempty"`
	Caption  string `json:"caption,omitempty"`
}

// InboundEvent represents an enriched incoming event forwarded to webhooks or dispatchers.
type InboundEvent struct {
	Event     string         `json:"event"` // e.g. "message.received"
	SessionID uint           `json:"session_id"`
	MessageID string         `json:"message_id"`
	Timestamp int64          `json:"timestamp"` // Unix timestamp in seconds
	Sender    SenderIdentity `json:"sender"`
	Message   InboundMessage `json:"message"`
}
