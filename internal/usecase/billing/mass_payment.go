package billing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainRouterJob "github.com/quixiq/polyglot/internal/domain/routerjob"
	domainTenant "github.com/quixiq/polyglot/internal/domain/tenant"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
	"github.com/quixiq/polyglot/pkg/logger"
)

// MassPaymentUseCase mengelola pembayaran massal tagihan via antrean job.
type MassPaymentUseCase struct {
	bulkRepo  port.BulkPaymentRepository
	invoices  port.InvoiceRepository
	processor port.PaymentProcessor
	jobs      port.RouterJobRepository
	reader    port.SettingReader
}

// NewMassPaymentUseCase creates a new MassPaymentUseCase.
func NewMassPaymentUseCase(
	bulkRepo port.BulkPaymentRepository,
	invoices port.InvoiceRepository,
	processor port.PaymentProcessor,
	jobs port.RouterJobRepository,
	reader port.SettingReader,
) *MassPaymentUseCase {
	return &MassPaymentUseCase{
		bulkRepo:  bulkRepo,
		invoices:  invoices,
		processor: processor,
		jobs:      jobs,
		reader:    reader,
	}
}

// CreateMassPayment creates a bulk payment batch and enqueues a background job.
func (u *MassPaymentUseCase) CreateMassPayment(ctx context.Context, tenantID string, createdBy *uint, paymentMethodID *string, invoiceIDs []string, notes string) (domainBilling.BulkPaymentBatch, error) {
	if len(invoiceIDs) == 0 {
		return domainBilling.BulkPaymentBatch{}, fmt.Errorf("%w: invoice_ids cannot be empty", domainBilling.ErrInvalidInput)
	}
	if tenantID == "" {
		tenantID = domainTenant.DefaultID
	}

	now := time.Now().UTC()
	batchID := idgen.New("bpb")
	batch := domainBilling.BulkPaymentBatch{
		ID:              batchID,
		TenantID:        tenantID,
		CreatedBy:       createdBy,
		PaymentMethodID: paymentMethodID,
		InvoiceCount:    len(invoiceIDs),
		Status:          domainBilling.BulkStatusQueued,
		Notes:           notes,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	items := make([]domainBilling.BulkPaymentItem, 0, len(invoiceIDs))
	for _, invID := range invoiceIDs {
		inv, err := u.invoices.FindByID(ctx, invID)
		if err != nil {
			return domainBilling.BulkPaymentBatch{}, fmt.Errorf("invoice %s: %w", invID, err)
		}
		outstanding := inv.Outstanding()
		if outstanding <= 0 {
			continue // skip already paid
		}
		items = append(items, domainBilling.BulkPaymentItem{
			ID:        idgen.New("bpi"),
			BatchID:   batchID,
			InvoiceID: inv.ID,
			Amount:    outstanding,
			Status:    domainBilling.BulkItemQueued,
			CreatedAt: now,
		})
	}

	if len(items) == 0 {
		return domainBilling.BulkPaymentBatch{}, fmt.Errorf("%w: all selected invoices are already paid", domainBilling.ErrInvoiceAlreadyPaid)
	}
	batch.InvoiceCount = len(items)

	if err := u.bulkRepo.SaveBatch(ctx, batch); err != nil {
		return domainBilling.BulkPaymentBatch{}, fmt.Errorf("save batch: %w", err)
	}
	if err := u.bulkRepo.SaveItems(ctx, items); err != nil {
		return domainBilling.BulkPaymentBatch{}, fmt.Errorf("save items: %w", err)
	}

	// Enqueue job router jika queue tersedia
	if u.jobs != nil {
		payload, _ := json.Marshal(map[string]any{"batch_id": batchID})
		job, err := u.jobs.Enqueue(ctx, domainRouterJob.Job{
			ID:             idgen.New("rj"),
			TenantID:       tenantID,
			DeviceID:       "system",
			JobType:        domainRouterJob.TypeMassPayment,
			Payload:        payload,
			IdempotencyKey: fmt.Sprintf("mass-pay:%s", batchID),
			Status:         domainRouterJob.StatusQueued,
			MaxAttempts:    3,
			NextRunAt:      now,
		})
		if err == nil {
			batch.JobID = job.ID
			_ = u.bulkRepo.SaveBatch(ctx, batch)
		} else {
			logger.WithComponent("MassPayment").WithError(err).Warn("enqueue mass payment job failed")
		}
	}

	return batch, nil
}

// ListMassPayments returns recent batches.
func (u *MassPaymentUseCase) ListMassPayments(ctx context.Context, tenantID string, limit int) ([]domainBilling.BulkPaymentBatch, error) {
	return u.bulkRepo.ListBatches(ctx, tenantID, limit)
}

// GetMassPayment retrieves batch and its items.
func (u *MassPaymentUseCase) GetMassPayment(ctx context.Context, batchID string) (domainBilling.BulkPaymentBatch, []domainBilling.BulkPaymentItem, error) {
	batch, err := u.bulkRepo.FindBatchByID(ctx, batchID)
	if err != nil {
		return domainBilling.BulkPaymentBatch{}, nil, err
	}
	items, err := u.bulkRepo.ListItems(ctx, batchID)
	if err != nil {
		return domainBilling.BulkPaymentBatch{}, nil, err
	}
	return batch, items, nil
}

// ProcessBatch executes items in a batch sequentially (called by background worker).
func (u *MassPaymentUseCase) ProcessBatch(ctx context.Context, batchID string) error {
	batch, err := u.bulkRepo.FindBatchByID(ctx, batchID)
	if err != nil {
		return err
	}

	_ = u.bulkRepo.UpdateBatchStatus(ctx, batchID, domainBilling.BulkStatusRunning)

	cashAccountID := "ca-1001-kas-kantor"
	incomeCategoryID := "cc-tagihan"
	if u.reader != nil {
		if v := u.reader.GetValue(ctx, "isp.manual_payment_cash_account_id", ""); v != "" {
			cashAccountID = v
		}
		if v := u.reader.GetValue(ctx, "isp.manual_payment_income_category_id", ""); v != "" {
			incomeCategoryID = v
		}
	}

	for {
		item, found, errClaim := u.bulkRepo.ClaimNextQueuedItem(ctx, batchID)
		if errClaim != nil {
			return errClaim
		}
		if !found {
			break
		}

		now := time.Now().UTC()
		pay, errPay := u.processor.ProcessCashPayment(ctx, port.CashPaymentCommand{
			TenantID:         batch.TenantID,
			InvoiceID:        item.InvoiceID,
			PaymentMethodID:  batch.PaymentMethodID,
			Amount:           item.Amount,
			ReceivedBy:       batch.CreatedBy,
			CashAccountID:    cashAccountID,
			IncomeCategoryID: incomeCategoryID,
			ScanMethod:       domainBilling.ScanManual,
			Notes:            fmt.Sprintf("Mass payment batch %s", batchID),
		})

		item.ProcessedAt = &now
		if errPay != nil {
			if errors.Is(errPay, domainBilling.ErrInvoiceAlreadyPaid) {
				item.Status = domainBilling.BulkItemSucceeded
			} else {
				item.Status = domainBilling.BulkItemFailed
				item.Error = errPay.Error()
			}
		} else {
			item.Status = domainBilling.BulkItemSucceeded
			item.PaymentID = &pay.ID
		}
		_ = u.bulkRepo.UpdateItem(ctx, item)
	}

	queued, processing, succeeded, failed, _ := u.bulkRepo.CountItems(ctx, batchID)
	if queued == 0 && processing == 0 {
		finalStatus := domainBilling.BulkStatusDone
		if failed > 0 && succeeded == 0 {
			finalStatus = domainBilling.BulkStatusFailed
		} else if failed > 0 && succeeded > 0 {
			finalStatus = domainBilling.BulkStatusPartial
		}
		_ = u.bulkRepo.UpdateBatchStatus(ctx, batchID, finalStatus)
	}

	return nil
}
