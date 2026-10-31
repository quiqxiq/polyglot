package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/billing"
)

// PaymentSubmissionFilter menyaring daftar bukti transfer manual.
type PaymentSubmissionFilter struct {
	Status string // PENDING | APPROVED | REJECTED; kosong = semua
	Limit  int
}

// PaymentSubmissionRepository menyimpan bukti transfer manual (Fase 3).
type PaymentSubmissionRepository interface {
	Save(ctx context.Context, s billing.PaymentSubmission) error
	FindByID(ctx context.Context, id string) (billing.PaymentSubmission, error)
	List(ctx context.Context, f PaymentSubmissionFilter) ([]billing.PaymentSubmission, error)
}
