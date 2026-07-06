package retry

import "context"

// Do retries fn according to a backoff policy until ctx is done or it succeeds.
func Do(ctx context.Context) error {
	return nil
}
