package whatsapp

import "github.com/quixiq/polyglot/pkg/fault"

// Sentinel errors for the WhatsApp domain.
var (
	ErrSessionNotFound    = fault.New(fault.KindNotFound, "whatsapp: session not found")
	ErrClientDisconnected = fault.New(fault.KindUnavailable, "whatsapp: client is disconnected")
	ErrEmptyRecipient     = fault.New(fault.KindInvalidInput, "whatsapp: recipient phone number or jid is required")
	ErrEmptyMessage       = fault.New(fault.KindInvalidInput, "whatsapp: message content cannot be empty")
	ErrInvalidPhoneNumber = fault.New(fault.KindInvalidInput, "whatsapp: invalid phone number format")
	ErrInvalidPresence    = fault.New(fault.KindInvalidInput, "whatsapp: invalid chat presence state")
	ErrEmptyMessageID     = fault.New(fault.KindInvalidInput, "whatsapp: message id is required")
	ErrInvalidMedia       = fault.New(fault.KindInvalidInput, "whatsapp: invalid or empty media payload")
)
