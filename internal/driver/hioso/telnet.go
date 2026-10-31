package hioso

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"
)

// Telnet control bytes (RFC 854).
const (
	iac  = 255
	dont = 254
	do   = 253
	wont = 252
	will = 251
)

var (
	// promptRe mencocokkan akhir prompt mode user/enable/config HiOSO:
	// "EPON>", "EPON#", "EPON(config)#", "EPON(epon_0/1)#", "EPON(pon-1/1)#".
	promptRe = regexp.MustCompile(`(?m)EPON[^\r\n]*[>#]\s*$`)
	// menuRe mendeteksi menu pemilihan mode manajemen (beberapa firmware).
	menuRe = regexp.MustCompile(`(?i)please choose`)
	// changeRe mendeteksi tawaran ganti password default.
	changeRe = regexp.MustCompile(`(?i)(change now|default password)`)
	// moreRe mendeteksi pagination CLI.
	moreRe = regexp.MustCompile(`(?i)(--\s*more\s*--|more\s*$)`)
	// loginRe mendeteksi prompt username/login.
	loginRe = regexp.MustCompile(`(?i)(user\s*name|username|login)\s*[:>]?\s*$`)
	// passwordRe mendeteksi prompt password.
	passwordRe = regexp.MustCompile(`(?i)password\s*[:>]?\s*$`)
)

// cliTransport adalah satu sesi CLI ke OLT (telnet/ssh) — dipisah agar CLI
// client bisa diuji dengan fake tanpa socket.
type cliTransport interface {
	Run(ctx context.Context, command string) (string, error)
	Close() error
}

// telnetTransport adalah sesi Telnet mentah ke OLT HiOSO: CRLF, filter IAC,
// login 3 lapis + menu + enable, dan penanganan pagination.
type telnetTransport struct {
	conn    net.Conn
	rd      *bufio.Reader
	timeout time.Duration
	host    string
}

// newTelnetTransport dial lalu login ke OLT.
func newTelnetTransport(ctx context.Context, host string, port int, username, password string, timeout time.Duration) (*telnetTransport, error) {
	if port <= 0 {
		port = 23
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	dialer := &net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("hioso: telnet dial %s: %w", addr, err)
	}
	t := &telnetTransport{conn: conn, rd: bufio.NewReaderSize(conn, 8192), timeout: timeout, host: host}
	if err := t.login(ctx, username, password); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return t, nil
}

func (t *telnetTransport) writeRaw(s string) error {
	_ = t.conn.SetWriteDeadline(time.Now().Add(t.timeout))
	_, err := t.conn.Write([]byte(s))
	return err
}

// writeLine menulis perintah dengan line ending CRLF (wajib untuk HiOSO).
func (t *telnetTransport) writeLine(s string) error {
	return t.writeRaw(strings.ReplaceAll(s, "\n", "\r\n") + "\r\n")
}

// login menjalankan urutan autentikasi yang bervariasi antar firmware:
// username/password → (menu mode "1") → (tolak ganti password) → enable.
func (t *telnetTransport) login(ctx context.Context, username, password string) error {
	out, _, err := t.readUntil(ctx, loginRe, passwordRe, promptRe, menuRe, changeRe)
	if err != nil {
		return fmt.Errorf("hioso: telnet banner: %w", err)
	}
	if loginRe.MatchString(out) && !passwordRe.MatchString(out) {
		if err := t.writeLine(username); err != nil {
			return err
		}
		out, _, err = t.readUntil(ctx, passwordRe, promptRe, menuRe)
		if err != nil {
			return fmt.Errorf("hioso: telnet password prompt: %w", err)
		}
	}
	if passwordRe.MatchString(out) {
		if err := t.writeLine(password); err != nil {
			return err
		}
	}

	// Fase pasca-password: bisa muncul menu mode, tawaran ganti password,
	// prompt enable (">"), atau langsung prompt "#".
	for attempt := 0; attempt < 6; attempt++ {
		out, _, err = t.readUntil(ctx, promptRe, menuRe, changeRe, passwordRe)
		if err != nil {
			return fmt.Errorf("hioso: telnet login: %w", err)
		}
		switch {
		case menuRe.MatchString(out) && !strings.Contains(out, "EPON#"):
			// "Please choose the management mode (1: CLI, ...)" → pilih CLI.
			if err := t.writeLine("1"); err != nil {
				return err
			}
			continue
		case changeRe.MatchString(out) && !strings.Contains(out, "EPON#"):
			// Tawaran ganti password default → tolak.
			if err := t.writeLine("N"); err != nil {
				return err
			}
			continue
		case passwordRe.MatchString(out) && !strings.Contains(out, "EPON#"):
			// Tier kedua (login → access) atau enable password.
			if err := t.writeLine(password); err != nil {
				return err
			}
			continue
		}
		break
	}

	last := lastLine(out)
	if strings.HasSuffix(strings.TrimSpace(last), ">") {
		if err := t.writeLine("enable"); err != nil {
			return err
		}
		out, _, err = t.readUntil(ctx, promptRe, passwordRe)
		if err != nil {
			return fmt.Errorf("hioso: telnet enable: %w", err)
		}
		if passwordRe.MatchString(out) {
			if err := t.writeLine(password); err != nil {
				return err
			}
			if _, _, err = t.readUntil(ctx, promptRe); err != nil {
				return fmt.Errorf("hioso: telnet enable password: %w", err)
			}
		}
	}
	return nil
}

// errPromptTimeout menandai habisnya deadline saat menunggu prompt.
var errPromptTimeout = errors.New("hioso: timeout menunggu prompt")

// readUntil membaca output sampai salah satu pattern cocok (nilai balik
// index pattern yang cocok) sambil memfilter IAC dan menjawab pagination.
func (t *telnetTransport) readUntil(ctx context.Context, patterns ...*regexp.Regexp) (string, int, error) {
	var sb strings.Builder
	tmp := make([]byte, 2048)
	iacState := make([]byte, 0, 3)
	for {
		if err := ctx.Err(); err != nil {
			return sb.String(), -1, err
		}
		deadline := time.Now().Add(t.timeout)
		if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
			deadline = d
		}
		_ = t.conn.SetReadDeadline(deadline)
		n, rerr := t.rd.Read(tmp)
		if n > 0 {
			clean := filterIAC(tmp[:n], &iacState, t)
			text := sb.String() + clean
			sb.Reset()
			sb.WriteString(text)
			if moreRe.MatchString(text) {
				// Halaman berikutnya: tekan Enter (didukung semua firmware HiOSO).
				_ = t.writeRaw("\r")
				sb.WriteString("\n")
				continue
			}
			for i, re := range patterns {
				if re.MatchString(text) {
					return text, i, nil
				}
			}
		}
		if rerr != nil {
			var netErr net.Error
			if errors.As(rerr, &netErr) && netErr.Timeout() {
				return sb.String(), -1, fmt.Errorf("%w (%s)", errPromptTimeout, t.host)
			}
			return sb.String(), -1, fmt.Errorf("hioso: telnet read %s: %w", t.host, rerr)
		}
	}
}

// Run mengirim satu perintah CLI dan mengembalikan output sampai prompt.
func (t *telnetTransport) Run(ctx context.Context, command string) (string, error) {
	if err := t.writeLine(command); err != nil {
		return "", fmt.Errorf("hioso: telnet write: %w", err)
	}
	out, _, err := t.readUntil(ctx, promptRe)
	if err != nil {
		return out, err
	}
	return stripEcho(out, command), nil
}

// Close menutup sesi (best effort berhenti ke user mode).
func (t *telnetTransport) Close() error {
	_ = t.writeLine("quit")
	return t.conn.Close()
}

// filterIAC membuang negosiasi Telnet dan menjawab DO/WILL agar sesi lanjut.
func filterIAC(in []byte, state *[]byte, t *telnetTransport) string {
	out := make([]byte, 0, len(in))
	for _, b := range in {
		switch len(*state) {
		case 0:
			if b == iac {
				*state = append(*state, b)
			} else {
				out = append(out, b)
			}
		case 1:
			switch b {
			case iac: // escaped 0xFF — data byte
				out = append(out, iac)
				*state = (*state)[:0]
			case will, wont, do, dont:
				*state = append(*state, b)
			default:
				*state = (*state)[:0]
			}
		case 2:
			// [IAC, CMD] + option byte.
			cmd, opt := (*state)[1], b
			switch cmd {
			case will, wont:
				_ = t.writeRaw(string([]byte{iac, dont, opt}))
			case do:
				_ = t.writeRaw(string([]byte{iac, wont, opt}))
			}
			*state = (*state)[:0]
		}
	}
	return string(out)
}

// stripEcho membuang baris pertama (echo perintah) dari output.
func stripEcho(out, command string) string {
	trimmed := strings.TrimLeft(out, "\r\n")
	firstNL := strings.IndexAny(trimmed, "\r\n")
	if firstNL >= 0 && strings.EqualFold(strings.TrimSpace(trimmed[:firstNL]), strings.TrimSpace(command)) {
		return strings.TrimLeft(trimmed[firstNL:], "\r\n")
	}
	return trimmed
}

func lastLine(s string) string {
	lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			return lines[i]
		}
	}
	return ""
}
