// Beban biaya gateway (planning v2 D11, analisis §6.1.B): MERCHANT vs
// CUSTOMER vs SPLIT + markup admin mitra.
package billing_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/port/mocktest"
	uc "github.com/quixiq/polyglot/internal/usecase/billing"
)

func TestCreateInvoice_MerchantBearsGatewayFee(t *testing.T) {
	invoices, customers, gwt, gateway, proc := gatewayFixture(t)
	settings := mocktest.NewFakeSettingReader(map[string]string{
		"isp.gateway_fee_bearer":    "MERCHANT",
		"isp.gateway_fee_percent":   "1",
		"isp.gateway_markup_amount": "500",
	})
	usecase := uc.NewGatewayChargeUseCase(invoices, customers, gwt, gateway, proc, settings)

	require.NoError(t, invoices.Save(context.Background(), unpaidInvoice("inv-merchant", "cust-m", "sub-m", 5)))
	require.NoError(t, customers.Save(context.Background(), customerWithPortal("cust-m", "99999999")))

	_, tx, err := usecase.CreateForInvoice(context.Background(), "inv-merchant", "", 60)
	require.NoError(t, err)

	// Markup tetap dibebankan ke pelanggan; biaya gateway provider ditanggung mitra.
	inv, err := invoices.FindByID(context.Background(), "inv-merchant")
	require.NoError(t, err)
	assert.InDelta(t, 110500, inv.Total, 0.01)
	assert.InDelta(t, 110500, tx.Amount, 0.01)
}

func TestCreateInvoice_CustomerBearsGatewayFee(t *testing.T) {
	invoices, customers, gwt, gateway, proc := gatewayFixture(t)
	settings := mocktest.NewFakeSettingReader(map[string]string{
		"isp.gateway_fee_bearer":    "CUSTOMER",
		"isp.gateway_fee_percent":   "1",
		"isp.gateway_markup_amount": "500",
	})
	usecase := uc.NewGatewayChargeUseCase(invoices, customers, gwt, gateway, proc, settings)

	require.NoError(t, invoices.Save(context.Background(), unpaidInvoice("inv-cust", "cust-c", "sub-c", 5)))
	require.NoError(t, customers.Save(context.Background(), customerWithPortal("cust-c", "99999999")))

	_, tx, err := usecase.CreateForInvoice(context.Background(), "inv-cust", "", 60)
	require.NoError(t, err)

	// fee 1% = 1100 + markup 500 = 1600 → total 111600.
	inv, err := invoices.FindByID(context.Background(), "inv-cust")
	require.NoError(t, err)
	assert.InDelta(t, 111600, inv.Total, 0.01)
	assert.InDelta(t, 111600, tx.Amount, 0.01)
}

func TestCreateInvoice_SplitBearsHalfGatewayFee(t *testing.T) {
	invoices, customers, gwt, gateway, proc := gatewayFixture(t)
	settings := mocktest.NewFakeSettingReader(map[string]string{
		"isp.gateway_fee_bearer":    "SPLIT",
		"isp.gateway_fee_percent":   "1",
		"isp.gateway_markup_amount": "0",
	})
	usecase := uc.NewGatewayChargeUseCase(invoices, customers, gwt, gateway, proc, settings)

	require.NoError(t, invoices.Save(context.Background(), unpaidInvoice("inv-split", "cust-s", "sub-s", 5)))
	require.NoError(t, customers.Save(context.Background(), customerWithPortal("cust-s", "99999999")))

	_, tx, err := usecase.CreateForInvoice(context.Background(), "inv-split", "", 60)
	require.NoError(t, err)
	// fee 1% = 1100, SPLIT = 550 → total 110550.
	assert.InDelta(t, 110550, tx.Amount, 0.01)
}

func TestAddGatewayFeeItem_Idempotent(t *testing.T) {
	invoices, _, _, _, _ := gatewayFixture(t)
	require.NoError(t, invoices.Save(context.Background(), unpaidInvoice("inv-idem", "cust-i", "sub-i", 5)))

	first, err := invoices.AddGatewayFeeItem(context.Background(), "inv-idem", 1500)
	require.NoError(t, err)
	assert.InDelta(t, 111500, first.Total, 0.01)

	second, err := invoices.AddGatewayFeeItem(context.Background(), "inv-idem", 1500)
	require.NoError(t, err)
	assert.InDelta(t, 111500, second.Total, 0.01, "fee tidak boleh dobel")
	assert.Len(t, invoices.ItemsOf("inv-idem"), 1)
}
