package whatsapp

import (
	domainBot "github.com/quixiq/polyglot/internal/domain/bot"
	domainWA "github.com/quixiq/polyglot/internal/domain/whatsapp"
)

// SessionDTO represents a WhatsApp session in HTTP responses.
type SessionDTO struct {
	ID          uint   `json:"id"`
	DeviceName  string `json:"device_name"`
	PhoneNumber string `json:"phone_number,omitempty"`
	Status      string `json:"status"`
	BotEnabled  bool   `json:"bot_enabled"`
}

// CreateSessionRequest represents input for creating a new session slot.
type CreateSessionRequest struct {
	DeviceName  string `json:"device_name"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

// PairingCodeRequest represents input for generating an 8-digit pairing code.
type PairingCodeRequest struct {
	SessionID   uint   `json:"session_id"`
	PhoneNumber string `json:"phone_number"`
}

// PairingCodeResponse returns the generated pairing code.
type PairingCodeResponse struct {
	SessionID   uint   `json:"session_id"`
	PairingCode string `json:"pairing_code"`
}

// SessionActionRequest represents common requests targeting a session ID.
type SessionActionRequest struct {
	SessionID uint `json:"session_id"`
}

// SendMessageRequest represents input for sending a plain text message.
type SendMessageRequest struct {
	SessionID uint   `json:"session_id"`
	To        string `json:"to"`
	Content   string `json:"content"`
}

// SendImageRequest represents input for sending an image with optional caption.
type SendImageRequest struct {
	SessionID   uint   `json:"session_id"`
	To          string `json:"to"`
	ImageBase64 string `json:"image_base64"`
	ContentType string `json:"content_type,omitempty"`
	Caption     string `json:"caption,omitempty"`
}

// SendDocumentRequest represents input for sending a file document with caption.
type SendDocumentRequest struct {
	SessionID      uint   `json:"session_id"`
	To             string `json:"to"`
	DocumentBase64 string `json:"document_base64"`
	FileName       string `json:"file_name"`
	ContentType    string `json:"content_type,omitempty"`
	Caption        string `json:"caption,omitempty"`
}

// ChatPresenceRequest represents input for updating typing / recording presence.
type ChatPresenceRequest struct {
	SessionID uint   `json:"session_id"`
	To        string `json:"to"`
	State     string `json:"state"` // "composing" or "paused"
	Media     string `json:"media"` // "text" or "audio"
}

// MarkReadRequest represents input for sending read receipts for messages.
type MarkReadRequest struct {
	SessionID  uint     `json:"session_id"`
	ChatJID    string   `json:"chat_jid"`
	SenderJID  string   `json:"sender_jid,omitempty"`
	MessageIDs []string `json:"message_ids"`
}

// GenericSuccessResponse represents a standard success envelope.
type GenericSuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

func toSessionDTO(s domainBot.WASession, status string) SessionDTO {
	return SessionDTO{
		ID:          s.ID,
		DeviceName:  s.DeviceName,
		PhoneNumber: s.PhoneNumber,
		Status:      status,
		BotEnabled:  s.IsBotEnabled,
	}
}

// IdentityResponse wraps the resolved sender identity in an HTTP response.
type IdentityResponse struct {
	Identity *domainWA.SenderIdentity `json:"identity"`
}
