package subscription

import "context"

// New creates a placeholder Subscription entity.
// Named New, not NewSubscription — CLAUDE.md §2.1 (avoid package name stutter).
// TODO: implement per Polyglot-Architecture.md business domain rules.
func New(ctx context.Context) error {
	return nil
}
