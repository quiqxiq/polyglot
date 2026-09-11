package port

import (
	"context"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	domainRegistration "github.com/quixiq/polyglot/internal/domain/registration"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
)

// ConversionArtifacts adalah agregat artefak konversi registrasi
// (INSTALLED → ACTIVE) yang harus dipersistensikan secara atomik.
type ConversionArtifacts struct {
	Registration domainRegistration.Registration
	Customer     domainCustomer.Customer
	Subscription domainSubscription.Subscription
	Invoice      domainBilling.Invoice
	Items        []domainBilling.InvoiceItem
}

// ConversionWriter mempersistensikan seluruh artefak konversi dalam SATU
// transaksi DB: customer, subscription (+enkripsi password), invoice + item,
// dan pembaruan registrasi (link artefak + status ACTIVE). Kegagalan di
// langkah mana pun me-rollback semuanya sehingga tidak ada artefak yatim.
type ConversionWriter interface {
	SaveConversion(ctx context.Context, artifacts ConversionArtifacts) error
}
