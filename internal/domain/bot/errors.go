package bot

import "github.com/quixiq/polyglot/pkg/fault"

// Sentinel errors for the bot domain: conversations, messages, and chat
// addressing.
var (
	// ErrConversationNotFound indicates the conversation does not exist.
	ErrConversationNotFound = fault.New(fault.KindNotFound, "bot: conversation not found")
	// ErrEmptyChatJID indicates a WhatsApp chat identifier is required.
	ErrEmptyChatJID = fault.New(fault.KindInvalidInput, "bot: chat_jid is required")
)
