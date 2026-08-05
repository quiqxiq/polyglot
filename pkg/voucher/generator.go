// Package voucher provides a template-based HTML renderer for Mikhmon-style
// hotspot voucher print sheets. It reads the header/row/footer templates from
// the internal/templates directory and fills each voucher row with the supplied
// data, including an inline base64-encoded QR code image.
//
// Supported layouts: "default", "small", "thermal".
package voucher

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	qrcode "github.com/skip2/go-qrcode"
)

// Layout defines the print size/format of voucher cards.
type Layout string

const (
	// LayoutDefault uses the standard card size (230px wide).
	LayoutDefault Layout = "default"
	// LayoutSmall uses a compact card size.
	LayoutSmall Layout = "small"
	// LayoutThermal uses a narrow thermal-printer format (180px wide).
	LayoutThermal Layout = "thermal"
)

// VoucherData holds the display values for a single voucher card.
// All fields map 1-to-1 with the placeholder tokens used in the template files.
type VoucherData struct {
	Username        string // %username%
	Password        string // %password%
	Price           string // %price%
	Validity        string // %validity%      — e.g. "1d", "7d"
	LimitUptime     string // %limitUptime%   — e.g. "24:00:00"
	LimitBytesTotal string // %limitBytesTotal%
	HotspotName     string // %hotspotName%
	DNSName         string // %dnsName%       — login URL host
	Logo            string // %logo%          — URL or base64 src
	Comment         string // %comment%       — displayed on thermal layout
	TimeStamp       string // %timeStamp%     — print timestamp
	Number          int    // %#%             — sequential card number
}

// QRSize is the pixel size of the generated QR code image (square).
const QRSize = 80

// generateQRBase64 encodes the given content as a QR code PNG and returns an
// HTML <img> tag with the image embedded as a base64 data URI.
// On error it returns an empty string (card renders without QR).
func generateQRBase64(content string) string {
	if content == "" {
		return ""
	}
	png, err := qrcode.Encode(content, qrcode.Medium, QRSize)
	if err != nil {
		return ""
	}
	b64 := base64.StdEncoding.EncodeToString(png)
	return fmt.Sprintf(`<img src="data:image/png;base64,%s" alt="QR" style="height:%dpx;width:%dpx;">`, b64, QRSize, QRSize)
}

// readTemplate reads a template file from the templateDir and returns its
// content as a string. Returns an error if the file cannot be read.
func readTemplate(templateDir, name string) (string, error) {
	path := filepath.Join(templateDir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("voucher: read template %q: %w", path, err)
	}
	return string(data), nil
}

// renderRow fills all placeholder tokens in a row template with voucher data.
func renderRow(rowTemplate string, v VoucherData, num int) string {
	// QR code encodes "username\npassword" so scanners show both values.
	qrContent := v.Username
	if v.Password != "" && v.Password != v.Username {
		qrContent = v.Username + "\n" + v.Password
	}
	qrTag := generateQRBase64(qrContent)

	ts := v.TimeStamp
	if ts == "" {
		ts = time.Now().Format("02/01/2006 15:04")
	}

	r := rowTemplate
	r = strings.ReplaceAll(r, "%username%", v.Username)
	r = strings.ReplaceAll(r, "%password%", v.Password)
	r = strings.ReplaceAll(r, "%price%", v.Price)
	r = strings.ReplaceAll(r, "%validity%", v.Validity)
	r = strings.ReplaceAll(r, "%limitUptime%", v.LimitUptime)
	r = strings.ReplaceAll(r, "%limitBytesTotal%", v.LimitBytesTotal)
	r = strings.ReplaceAll(r, "%hotspotName%", v.HotspotName)
	r = strings.ReplaceAll(r, "%dnsName%", v.DNSName)
	r = strings.ReplaceAll(r, "%logo%", v.Logo)
	r = strings.ReplaceAll(r, "%comment%", v.Comment)
	r = strings.ReplaceAll(r, "%timeStamp%", ts)
	r = strings.ReplaceAll(r, "%qrCode%", qrTag)
	r = strings.ReplaceAll(r, "%#%", fmt.Sprintf("%d", num))
	return r
}

// Render assembles a complete printable HTML page for the given vouchers.
// layout selects which set of template files to use ("default", "small", "thermal").
// templateDir is the path to the directory containing the *.txt template files
// (e.g. "internal/templates"). An absolute path or a path relative to the
// working directory are both accepted.
//
// The returned string is a self-contained HTML document ready to be sent as
// the body of an HTTP response with Content-Type: text/html; charset=utf-8.
func Render(vouchers []VoucherData, layout Layout, templateDir string) (string, error) {
	if layout == "" {
		layout = LayoutDefault
	}
	l := string(layout)

	header, err := readTemplate(templateDir, "header."+l+".txt")
	if err != nil {
		return "", err
	}
	rowTpl, err := readTemplate(templateDir, "row."+l+".txt")
	if err != nil {
		return "", err
	}
	footer, err := readTemplate(templateDir, "footer."+l+".txt")
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(header)
	for i, v := range vouchers {
		sb.WriteString(renderRow(rowTpl, v, i+1))
	}
	sb.WriteString(footer)
	return sb.String(), nil
}
