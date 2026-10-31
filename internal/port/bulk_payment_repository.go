package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/billing"
)

// BulkPaymentRepository mengelola batch mass payment beserta item-nya
// (progress dipersist agar job aman di-retry) — W3 PLAN-BILLING-OPS-V3.
type BulkPaymentRepository interface {
	SaveBatch(ctx context.Context, b billing.BulkPaymentBatch) error
	FindBatchByID(ctx context.Context, id string) (billing.BulkPaymentBatch, error)
	ListBatches(ctx context.Context, tenantID string, limit int) ([]billing.BulkPaymentBatch, error)
	UpdateBatchStatus(ctx context.Context, id, status string) error
	SaveItems(ctx context.Context, items []billing.BulkPaymentItem) error
	ListItems(ctx context.Context, batchID string) ([]billing.BulkPaymentItem, error)
	// ClaimNextQueuedItem mengambil satu item QUEUED dan menandainya PROCESSING
	// secara atomik; ok=false bila tidak ada item tersisa.
	ClaimNextQueuedItem(ctx context.Context, batchID string) (billing.BulkPaymentItem, bool, error)
	UpdateItem(ctx context.Context, item billing.BulkPaymentItem) error
	CountItems(ctx context.Context, batchID string) (queued, processing, succeeded, failed int, err error)
}
