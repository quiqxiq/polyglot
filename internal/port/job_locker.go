package port

import "context"

// JobLocker menyediakan mutual exclusion antar-instance untuk job periodik
// (mis. saat server di-scale horizontal) — F6-1. Implementasi default
// memakai PostgreSQL advisory lock; dialect lain selalu sukses.
type JobLocker interface {
	// TryLock mencoba mengambil lock bernama key. Fungsi unlock selalu
	// non-nil. ok=false berarti instance lain sedang memegang lock.
	TryLock(ctx context.Context, key string) (unlock func(), ok bool, err error)
}
