package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/port"
)

// ConversionWriter mempersistensikan artefak konversi registrasi secara
// atomik dalam satu transaksi PostgreSQL (F1-12).
type ConversionWriter struct {
	db    *gorm.DB
	vault port.CredentialVault
}

var _ port.ConversionWriter = (*ConversionWriter)(nil)

// NewConversionWriter constructs an atomic registration conversion writer.
func NewConversionWriter(db *gorm.DB, vault port.CredentialVault) *ConversionWriter {
	return &ConversionWriter{db: db, vault: vault}
}

// SaveConversion menulis customer, subscription (dengan password
// terenkripsi), invoice + item, dan registrasi dalam satu transaksi.
// Kegagalan di langkah mana pun me-rollback seluruh artefak.
func (w *ConversionWriter) SaveConversion(ctx context.Context, a port.ConversionArtifacts) error {
	return w.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(model.CustomerModelFromDomain(a.Customer)).Error; err != nil {
			return fmt.Errorf("save customer: %w", err)
		}

		subModel := model.SubscriptionModelFromDomain(a.Subscription)
		if w.vault != nil && a.Subscription.RemotePassword != "" {
			cipher, err := w.vault.EncryptString(ctx, a.Subscription.RemotePassword)
			if err != nil {
				return fmt.Errorf("encrypt subscription password: %w", err)
			}
			subModel.RemotePassword = cipher
		}
		if err := tx.Save(subModel).Error; err != nil {
			return fmt.Errorf("save subscription: %w", err)
		}

		if err := tx.Save(model.InvoiceModelFromDomain(a.Invoice)).Error; err != nil {
			return fmt.Errorf("save invoice: %w", err)
		}
		if len(a.Items) > 0 {
			items := make([]model.InvoiceItemModel, len(a.Items))
			for i := range a.Items {
				items[i] = *model.InvoiceItemModelFromDomain(a.Items[i])
			}
			if err := tx.Create(&items).Error; err != nil {
				return fmt.Errorf("save invoice items: %w", err)
			}
		}

		if err := tx.Save(model.RegistrationModelFromDomain(a.Registration)).Error; err != nil {
			return fmt.Errorf("save registration: %w", err)
		}
		return nil
	})
}
