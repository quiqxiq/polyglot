// F4-1/F4-6: registry multi-gateway, pemilihan default, dan guard transaksi
// PENDING ganda per invoice.
package billing_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	uc "github.com/quixiq/polyglot/internal/usecase/billing"
)

// namedGateway membungkus fake gateway dengan nama vendor yang bisa diatur.
type namedGateway struct {
	*mocktest.FakePaymentGateway
	name string
}

func (g namedGateway) Name() string { return g.name }

func TestGatewayRegistry_DefaultAndLookup(t *testing.T) {
	disabled := namedGateway{&mocktest.FakePaymentGateway{IsEnabled: false}, "TRIPAY"}
	enabled := namedGateway{&mocktest.FakePaymentGateway{IsEnabled: true}, "MIDTRANS"}

	// gw.active menunjuk gateway enabled → dipakai.
	reader := mocktest.NewFakeSettingReader(map[string]string{"gw.active": "MIDTRANS"})
	reg := uc.NewGatewayRegistry(reader, disabled, enabled)
	gw, err := reg.Default(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "MIDTRANS", gw.Name())

	// gw.active menunjuk gateway disabled → fallback yang enabled.
	reader2 := mocktest.NewFakeSettingReader(map[string]string{"gw.active": "TRIPAY"})
	reg2 := uc.NewGatewayRegistry(reader2, disabled, enabled)
	gw2, err := reg2.Default(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "MIDTRANS", gw2.Name())

	// Tidak ada yang enabled.
	reg3 := uc.NewGatewayRegistry(reader2, disabled)
	_, err = reg3.Default(context.Background())
	assert.ErrorIs(t, err, domainBilling.ErrGatewayDisabled)

	// Nama gateway tak dikenal.
	_, err = reg2.Get(context.Background(), "XENDIT")
	assert.ErrorIs(t, err, domainBilling.ErrGatewayNotFound)

	// Lookup case-insensitive.
	got, err := reg2.Get(context.Background(), "midtrans")
	require.NoError(t, err)
	assert.Equal(t, "MIDTRANS", got.Name())
}

func TestCreateForInvoiceVia_ReusesPendingAndExpiresOtherGateway(t *testing.T) {
	invoices := mocktest.NewFakeInvoiceRepo()
	customers := mocktest.NewFakeCustomerRepo()
	gwt := mocktest.NewFakeGatewayTxRepo()
	proc := &FakeProcessorRecorder{}
	reader := mocktest.NewFakeSettingReader(map[string]string{"gw.active": "FAKE"})

	primary := namedGateway{&mocktest.FakePaymentGateway{
		IsEnabled: true,
		Charge: port.ChargeResult{
			ExternalID: "EXT-1", PaymentURL: "https://pay/1", Status: "PENDING",
		},
	}, "FAKE"}
	secondary := namedGateway{&mocktest.FakePaymentGateway{
		IsEnabled: true,
		Charge: port.ChargeResult{
			ExternalID: "EXT-2", PaymentURL: "https://pay/2", Status: "PENDING",
		},
	}, "OTHER"}
	reg := uc.NewGatewayRegistry(reader, primary, secondary)
	usecase := uc.NewGatewayChargeUseCaseWithRegistry(invoices, customers, gwt, reg, proc, reader)

	ctx := context.Background()
	require.NoError(t, invoices.Save(ctx, unpaidInvoice("inv-multi", "cust-gw", "sub-gw", 5)))
	require.NoError(t, customers.Save(ctx, customerWithPortal("cust-gw", "99999999")))

	// Charge pertama membuat transaksi.
	res1, tx1, err := usecase.CreateForInvoice(ctx, "inv-multi", "", 60)
	require.NoError(t, err)
	assert.Equal(t, "EXT-1", res1.ExternalID)

	// Charge kedua gateway sama → reuse transaksi PENDING (idempoten).
	res2, tx2, err := usecase.CreateForInvoice(ctx, "inv-multi", "", 60)
	require.NoError(t, err)
	assert.Equal(t, tx1.ID, tx2.ID)
	assert.Equal(t, "EXT-1", res2.ExternalID)
	assert.Equal(t, "https://pay/1", res2.PaymentURL)

	// Pindah gateway → transaksi lama di-expire, transaksi baru dibuat.
	res3, tx3, err := usecase.CreateForInvoiceVia(ctx, "inv-multi", "OTHER", "", 60)
	require.NoError(t, err)
	assert.Equal(t, "EXT-2", res3.ExternalID)
	assert.NotEqual(t, tx1.ID, tx3.ID)

	txs, err := gwt.FindByInvoice(ctx, "inv-multi")
	require.NoError(t, err)
	require.Len(t, txs, 2)
	statuses := map[string]string{}
	for _, tx := range txs {
		statuses[tx.Gateway] = tx.Status
	}
	assert.Equal(t, domainBilling.GatewayStatusExpired, statuses["FAKE"])
	assert.Equal(t, domainBilling.GatewayStatusPending, statuses["OTHER"])
}
