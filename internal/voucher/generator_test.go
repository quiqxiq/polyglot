package voucher_test

import (
	"strings"
	"testing"
	"testing/fstest"

	templatefs "github.com/quixiq/polyglot/internal/template"
	"github.com/quixiq/polyglot/internal/voucher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// templates is the embedded template FS used by all render tests.
var templates = templatefs.FS

func sampleVoucher(n int) voucher.VoucherData {
	return voucher.VoucherData{
		Username:        "user123",
		Password:        "pass456",
		Price:           "10000",
		Validity:        "1d",
		LimitUptime:     "",
		LimitBytesTotal: "",
		HotspotName:     "TestHotspot",
		DNSName:         "192.168.1.1",
		Logo:            "",
		Comment:         "vc-A1B-08.05.26",
		TimeStamp:       "05/08/2026 10:00",
		Number:          n,
	}
}

func TestRender_DefaultLayout(t *testing.T) {
	v := sampleVoucher(1)

	html, err := voucher.Render([]voucher.VoucherData{v}, voucher.LayoutDefault, templates)
	require.NoError(t, err)

	assert.Contains(t, html, "user123", "HTML should contain username")
	assert.Contains(t, html, "pass456", "HTML should contain password")
	assert.Contains(t, html, "10000", "HTML should contain price")
	assert.Contains(t, html, "TestHotspot", "HTML should contain hotspot name")
	assert.True(t, strings.HasPrefix(strings.TrimSpace(html), "<!DOCTYPE html>"), "HTML should start with DOCTYPE")
	assert.Contains(t, html, "</html>", "HTML should end with </html>")
}

func TestRender_SmallLayout(t *testing.T) {
	v := sampleVoucher(1)

	html, err := voucher.Render([]voucher.VoucherData{v}, voucher.LayoutSmall, templates)
	require.NoError(t, err)

	assert.Contains(t, html, "user123")
	assert.Contains(t, html, "</html>")
}

func TestRender_ThermalLayout(t *testing.T) {
	v := sampleVoucher(1)

	html, err := voucher.Render([]voucher.VoucherData{v}, voucher.LayoutThermal, templates)
	require.NoError(t, err)

	assert.Contains(t, html, "user123")
	assert.Contains(t, html, "pass456")
	assert.Contains(t, html, "</html>")
}

func TestRender_MultiplVouchers(t *testing.T) {
	vouchers := []voucher.VoucherData{
		{Username: "a1b2c3", Password: "a1b2c3", Number: 1, HotspotName: "HS", DNSName: "192.168.1.1"},
		{Username: "d4e5f6", Password: "xyz789", Number: 2, HotspotName: "HS", DNSName: "192.168.1.1"},
	}

	html, err := voucher.Render(vouchers, voucher.LayoutDefault, templates)
	require.NoError(t, err)

	assert.Contains(t, html, "a1b2c3", "first voucher username should appear")
	assert.Contains(t, html, "d4e5f6", "second voucher username should appear")
	assert.Contains(t, html, "xyz789", "second voucher password should appear")
}

func TestRender_QRCodeEmbedded(t *testing.T) {
	v := sampleVoucher(1)

	html, err := voucher.Render([]voucher.VoucherData{v}, voucher.LayoutDefault, templates)
	require.NoError(t, err)

	// QR code should be embedded as base64 data URI in an <img> tag.
	assert.Contains(t, html, "data:image/png;base64,", "QR code should be inline base64")
	assert.Contains(t, html, `<img src="data:image/png;base64,`, "QR code should be in an img tag")
}

func TestRender_EmptyLayout_FallsBackToDefault(t *testing.T) {
	v := sampleVoucher(1)

	html, err := voucher.Render([]voucher.VoucherData{v}, "", templates)
	require.NoError(t, err)
	assert.Contains(t, html, "user123")
}

func TestRender_InvalidLayout_ReturnsError(t *testing.T) {
	v := sampleVoucher(1)

	_, err := voucher.Render([]voucher.VoucherData{v}, "nonexistent", templates)
	assert.Error(t, err, "invalid layout should return error")
}

func TestRender_MissingTemplateFS_ReturnsError(t *testing.T) {
	v := sampleVoucher(1)
	_, err := voucher.Render([]voucher.VoucherData{v}, voucher.LayoutDefault, fstest.MapFS{})
	assert.Error(t, err)
}

func TestQRContent_CredentialsMode(t *testing.T) {
	v := voucher.VoucherData{Username: "user123", Password: "pass456"}
	assert.Equal(t, "user123\npass456", voucher.QRContent(v, voucher.QRModeCredentials))

	// Password same as username → single line (existing behavior).
	same := voucher.VoucherData{Username: "a1b2c3", Password: "a1b2c3"}
	assert.Equal(t, "a1b2c3", voucher.QRContent(same, voucher.QRModeCredentials))
}

func TestQRContent_LoginURLMode(t *testing.T) {
	v := voucher.VoucherData{Username: "user 1", Password: "p@ss", DNSName: "192.168.1.1"}
	got := voucher.QRContent(v, voucher.QRModeLoginURL)
	assert.Equal(t, "http://192.168.1.1/login?username=user+1&password=p%40ss", got)
}

func TestRenderWithOptions_LoginURLQR(t *testing.T) {
	v := sampleVoucher(1)
	v.DNSName = "192.168.1.1"

	html, err := voucher.RenderWithOptions([]voucher.VoucherData{v}, voucher.LayoutDefault, templates, voucher.Options{QRMode: voucher.QRModeLoginURL})
	require.NoError(t, err)

	assert.Contains(t, html, "data:image/png;base64,", "QR code should still be embedded as base64")
	assert.Contains(t, html, "192.168.1.1", "DNS name used for the login URL should appear in HTML")
}

func TestListTemplates(t *testing.T) {
	infos := voucher.ListTemplates()
	require.Len(t, infos, 3)
	assert.Equal(t, "default", infos[0].Name)
	assert.Equal(t, []string{"header", "row", "footer"}, infos[0].Sections)
	assert.Equal(t, "small", infos[1].Name)
	assert.Equal(t, "thermal", infos[2].Name)
}

func TestTemplateFile(t *testing.T) {
	assert.Equal(t, "header.default.txt", voucher.TemplateFile("default", "header"))
	assert.Equal(t, "footer.thermal.txt", voucher.TemplateFile("thermal", "footer"))
}

func TestRender_SequentialNumbers(t *testing.T) {
	vouchers := []voucher.VoucherData{
		{Username: "u1", Password: "u1", Number: 1, HotspotName: "HS", DNSName: "192.168.1.1"},
		{Username: "u2", Password: "u2", Number: 2, HotspotName: "HS", DNSName: "192.168.1.1"},
		{Username: "u3", Password: "u3", Number: 3, HotspotName: "HS", DNSName: "192.168.1.1"},
	}

	html, err := voucher.Render(vouchers, voucher.LayoutDefault, templates)
	require.NoError(t, err)

	// Sequential numbers [%#%] should appear in output.
	assert.Contains(t, html, "[1]")
	assert.Contains(t, html, "[2]")
	assert.Contains(t, html, "[3]")
}
