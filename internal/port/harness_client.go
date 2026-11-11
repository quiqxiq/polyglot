package port

import "context"

// HarnessClient executes one one-shot DeepSeek harness turn.
//
// Implementations must run the harness with a read-only profile, pass the
// prompt over stdin rather than argv, and return only the committed final
// assistant text. Errors are transport or run failures; callers must not send
// their details to end users.
type HarnessClient interface {
	Ask(ctx context.Context, prompt string) (string, error)
}
