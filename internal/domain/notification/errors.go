package notification

import "github.com/quixiq/polyglot/pkg/fault"

// Sentinel errors for the notification domain.
var (
	// ErrNoWASession indicates no connected WhatsApp session exists for the
	// tenant, so the notification cannot be delivered.
	ErrNoWASession = fault.New(fault.KindFailedPrecondition, "notification: no connected whatsapp session")
)
