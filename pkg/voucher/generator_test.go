package voucher_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/quixiq/polyglot/pkg/voucher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// templateDir returns the absolute path to internal/templates relative to this
// test file's location (pkg/voucher/ → ../../internal/templates).
func templateDir(t *testing.T) string {
	t.Helper()
	// Resolve from pkg/voucher to project root then into internal/templates.
	cwd, err := os.Getwd()
	require.NoError(t, err)
	return filepath.Join(cwd, "..", "..", "internal", "templates")
}

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
	tdir := templateDir(t)
	v := sampleVoucher(1)

	html, err := voucher.Render([]voucher.VoucherData{v}, voucher.LayoutDefault, tdir)
	require.NoError(t, err)

	assert.Contains(t, html, "user123", "HTML should contain username")
	assert.Contains(t, html, "pass456", "HTML should contain password")
	assert.Contains(t, html, "10000", "HTML should contain price")
	assert.Contains(t, html, "TestHotspot", "HTML should contain hotspot name")
	assert.True(t, strings.HasPrefix(strings.TrimSpace(html), "<!DOCTYPE html>"), "HTML should start with DOCTYPE")
	assert.Contains(t, html, "</html>", "HTML should end with </html>")
}

func TestRender_SmallLayout(t *testing.T) {
	tdir := templateDir(t)
	v := sampleVoucher(1)

	html, err := voucher.Render([]voucher.VoucherData{v}, voucher.LayoutSmall, tdir)
	require.NoError(t, err)

	assert.Contains(t, html, "user123")
	assert.Contains(t, html, "</html>")
}

func TestRender_ThermalLayout(t *testing.T) {
	tdir := templateDir(t)
	v := sampleVoucher(1)

	html, err := voucher.Render([]voucher.VoucherData{v}, voucher.LayoutThermal, tdir)
	require.NoError(t, err)

	assert.Contains(t, html, "user123")
	assert.Contains(t, html, "pass456")
	assert.Contains(t, html, "</html>")
}

func TestRender_MultiplVouchers(t *testing.T) {
	tdir := templateDir(t)
	vouchers := []voucher.VoucherData{
		{Username: "a1b2c3", Password: "a1b2c3", Number: 1, HotspotName: "HS", DNSName: "192.168.1.1"},
		{Username: "d4e5f6", Password: "xyz789", Number: 2, HotspotName: "HS", DNSName: "192.168.1.1"},
	}

	html, err := voucher.Render(vouchers, voucher.LayoutDefault, tdir)
	require.NoError(t, err)

	assert.Contains(t, html, "a1b2c3", "first voucher username should appear")
	assert.Contains(t, html, "d4e5f6", "second voucher username should appear")
	assert.Contains(t, html, "xyz789", "second voucher password should appear")
}

func TestRender_QRCodeEmbedded(t *testing.T) {
	tdir := templateDir(t)
	v := sampleVoucher(1)

	html, err := voucher.Render([]voucher.VoucherData{v}, voucher.LayoutDefault, tdir)
	require.NoError(t, err)

	// QR code should be embedded as base64 data URI in an <img> tag.
	assert.Contains(t, html, "data:image/png;base64,", "QR code should be inline base64")
	assert.Contains(t, html, `<img src="data:image/png;base64,`, "QR code should be in an img tag")
}

func TestRender_EmptyLayout_FallsBackToDefault(t *testing.T) {
	tdir := templateDir(t)
	v := sampleVoucher(1)

	html, err := voucher.Render([]voucher.VoucherData{v}, "", tdir)
	require.NoError(t, err)
	assert.Contains(t, html, "user123")
}

func TestRender_InvalidLayout_ReturnsError(t *testing.T) {
	tdir := templateDir(t)
	v := sampleVoucher(1)

	_, err := voucher.Render([]voucher.VoucherData{v}, "nonexistent", tdir)
	assert.Error(t, err, "invalid layout should return error")
}

func TestRender_InvalidTemplateDir_ReturnsError(t *testing.T) {
	v := sampleVoucher(1)
	_, err := voucher.Render([]voucher.VoucherData{v}, voucher.LayoutDefault, "/nonexistent/path")
	assert.Error(t, err)
}

func TestRender_SequentialNumbers(t *testing.T) {
	tdir := templateDir(t)
	vouchers := []voucher.VoucherData{
		{Username: "u1", Password: "u1", Number: 1, HotspotName: "HS", DNSName: "192.168.1.1"},
		{Username: "u2", Password: "u2", Number: 2, HotspotName: "HS", DNSName: "192.168.1.1"},
		{Username: "u3", Password: "u3", Number: 3, HotspotName: "HS", DNSName: "192.168.1.1"},
	}

	html, err := voucher.Render(vouchers, voucher.LayoutDefault, tdir)
	require.NoError(t, err)

	// Sequential numbers [%#%] should appear in output.
	assert.Contains(t, html, "[1]")
	assert.Contains(t, html, "[2]")
	assert.Contains(t, html, "[3]")
}
