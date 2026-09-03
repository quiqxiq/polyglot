package bot

import "github.com/quixiq/polyglot/pkg/fault"

// Sentinel errors for the bot domain: conversations, messages, and chat
// addressing.
var (
	// ErrConversationNotFound indicates the conversation does not exist.
	ErrConversationNotFound = fault.New(fault.KindNotFound, "bot: conversation not found")
	// ErrNotFound indicates the requested bot entity was not found.
	ErrNotFound = fault.New(fault.KindNotFound, "bot: record not found")
	// ErrEmptyChatJID indicates a WhatsApp chat identifier is required.
	ErrEmptyChatJID = fault.New(fault.KindInvalidInput, "bot: chat_jid is required")
	// ErrConversationIDRequired indicates a missing conversation identifier.
	ErrConversationIDRequired = fault.New(fault.KindInvalidInput, "bot: conversation id is required")
	// ErrGatewayNotInitialized indicates the engine was wired without a
	// WhatsApp gateway (composition-root error).
	ErrGatewayNotInitialized = fault.New(fault.KindFailedPrecondition, "bot: whatsapp gateway not initialized")
	// ErrConversationServiceNotInitialized indicates the engine was wired
	// without a conversation service (composition-root error).
	ErrConversationServiceNotInitialized = fault.New(fault.KindFailedPrecondition, "bot: conversation service not initialized")
)
