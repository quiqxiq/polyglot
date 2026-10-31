package port

import (
	"context"
	"time"

	"github.com/quixiq/polyglot/internal/domain/approval"
)

// ApprovalFilter narrows approval listings.
type ApprovalFilter struct {
	TenantID string
	DeviceID string
	Status   string
	Limit    int
	Offset   int
}

// ApprovalRepository persists approval_requests. Decisions and execution use
// compare-and-set on (revision, status) so two humans/devices cannot race.
type ApprovalRepository interface {
	Create(ctx context.Context, req approval.Request) error
	FindByID(ctx context.Context, id string) (approval.Request, error)
	FindByIdempotencyKey(ctx context.Context, tenantID, key string) (approval.Request, error)
	List(ctx context.Context, f ApprovalFilter) ([]approval.Request, int, error)
	// Update saves req only when the stored revision equals expectedRevision.
	// It returns false (and no error) on a CAS mismatch.
	UpdateIfRevision(ctx context.Context, req approval.Request, expectedRevision int) (bool, error)
	// ClaimForExecution atomically moves APPROVED→EXECUTING with a row lock
	// and returns the claimed request. It returns false when another actor
	// already claimed it, it expired, or the status was not APPROVED.
	ClaimForExecution(ctx context.Context, id string, now time.Time) (approval.Request, bool, error)
	// ExpirePending marks non-decided PENDING approvals past expires_at as EXPIRED.
	ExpirePending(ctx context.Context, now time.Time) (int, error)
}
