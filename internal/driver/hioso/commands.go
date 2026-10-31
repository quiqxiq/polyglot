package hioso

import (
	"context"
	"fmt"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/olt"
	"github.com/quixiq/polyglot/internal/port"
)

// CLIClient adalah perintah terstruktur HiOSO di atas satu sesi CLI.
type CLIClient struct {
	t cliTransport
}

func newCLIClient(t cliTransport) *CLIClient {
	return &CLIClient{t: t}
}

// ListONUs membaca inventori via CLI (fallback bila SNMP gagal). Untuk
// HA7302 satu domain LLID datar; untuk HA7304 per PON.
func (c *CLIClient) ListONUs(ctx context.Context, caps olt.Capabilities, ponCount int) ([]port.OLTONU, error) {
	if ponCount <= 0 {
		ponCount = caps.PONCount
	}
	pons := ponCount
	if caps.FlatLLID {
		pons = 1
	}
	var rows []port.OLTONU
	for pon := 1; pon <= pons; pon++ {
		out, err := c.run(ctx, fmt.Sprintf("show onu info epon 0/%d all", pon))
		if err != nil {
			return nil, err
		}
		rows = append(rows, parseShowONUInfo(out, caps.Variant)...)
	}
	return rows, nil
}

// ShowVersion mengembalikan output `show version`.
func (c *CLIClient) ShowVersion(ctx context.Context) (string, error) {
	return c.run(ctx, "show version")
}

// ShowRunningConfig mengembalikan konfigurasi berjalan OLT.
func (c *CLIClient) ShowRunningConfig(ctx context.Context) (string, error) {
	return c.run(ctx, "show running-config")
}

// SystemInfo menghasilkan key-value dari `show version` + `show system`.
func (c *CLIClient) SystemInfo(ctx context.Context) (map[string]string, error) {
	out, err := c.run(ctx, "show version")
	if err != nil {
		return nil, err
	}
	info := parseKeyValue(out)
	if sys, errSys := c.run(ctx, "show system"); errSys == nil {
		for k, v := range parseKeyValue(sys) {
			if _, exists := info[k]; !exists {
				info[k] = v
			}
		}
	}
	return info, nil
}

// RenameONU mengganti nama ONU. HA7302 memakai SNMP SET (ditangani driver),
// HA7304 memakai CLI di interface PON.
func (c *CLIClient) RenameONU(ctx context.Context, caps olt.Capabilities, ponIndex, onuIndex int, name string) error {
	if caps.RenameViaSNMP {
		return fmt.Errorf("%w: rename HA7302 via SNMP SET", olt.ErrUnsupportedVariant)
	}
	if err := olt.ValidateONUName(name); err != nil {
		return err
	}
	if err := c.enterPONInterface(ctx, ponIndex); err != nil {
		return err
	}
	defer c.exitToPrivileged(ctx)
	return c.exec(ctx, fmt.Sprintf("onu %d name %s", onuIndex, name))
}

// RebootONU me-reboot ONU.
func (c *CLIClient) RebootONU(ctx context.Context, caps olt.Capabilities, ponIndex, onuIndex int) error {
	if caps.FlatLLID {
		// HA7302: epon → pon 1/{PON} → set onu {N} reboot.
		if _, err := c.run(ctx, "epon"); err != nil {
			return err
		}
		if _, err := c.run(ctx, fmt.Sprintf("pon 1/%d", ponIndex)); err != nil {
			return err
		}
		defer c.exitToPrivileged(ctx)
		return c.exec(ctx, fmt.Sprintf("set onu %d reboot", onuIndex))
	}
	if err := c.enterPONInterface(ctx, ponIndex); err != nil {
		return err
	}
	defer c.exitToPrivileged(ctx)
	return c.exec(ctx, fmt.Sprintf("onu %d reboot", onuIndex))
}

// SetONUAdmin mengaktifkan/menonaktifkan ONU.
func (c *CLIClient) SetONUAdmin(ctx context.Context, caps olt.Capabilities, ponIndex, onuIndex int, enabled bool) error {
	action := "deactivate"
	if enabled {
		action = "activate"
	}
	if caps.FlatLLID {
		// TODO(phase0): verifikasi dialek auth-mode HA7302 pada perangkat nyata.
		if _, err := c.run(ctx, "epon"); err != nil {
			return err
		}
		if _, err := c.run(ctx, fmt.Sprintf("pon 1/%d", ponIndex)); err != nil {
			return err
		}
		defer c.exitToPrivileged(ctx)
		mode := "pass"
		if !enabled {
			mode = "deny"
		}
		return c.exec(ctx, fmt.Sprintf("set onu %d auth-mode %s", onuIndex, mode))
	}
	if err := c.enterPONInterface(ctx, ponIndex); err != nil {
		return err
	}
	defer c.exitToPrivileged(ctx)
	return c.exec(ctx, fmt.Sprintf("onu %d %s", onuIndex, action))
}

// DeleteONU menghapus ONU dari OLT.
func (c *CLIClient) DeleteONU(ctx context.Context, caps olt.Capabilities, ponIndex, onuIndex int) error {
	if caps.FlatLLID {
		return c.exec(ctx, fmt.Sprintf("delete onu 1/%d/%d", ponIndex, onuIndex))
	}
	if err := c.enterPONInterface(ctx, ponIndex); err != nil {
		return err
	}
	defer c.exitToPrivileged(ctx)
	return c.exec(ctx, fmt.Sprintf("delete onu %d", onuIndex))
}

// SaveConfig menyimpan konfigurasi ke NVRAM.
func (c *CLIClient) SaveConfig(ctx context.Context) error {
	out, err := c.run(ctx, "write file")
	if err != nil {
		return err
	}
	if commandRejected(out) {
		// Beberapa firmware hanya mengenal "write".
		out, err = c.run(ctx, "write")
		if err != nil {
			return err
		}
	}
	if commandRejected(out) {
		return fmt.Errorf("%w: %s", olt.ErrCommandRejected, firstMeaningfulLine(out))
	}
	return nil
}

// --- helper internal ---

func (c *CLIClient) run(ctx context.Context, command string) (string, error) {
	out, err := c.t.Run(ctx, command)
	if err != nil {
		return out, err
	}
	return out, nil
}

func (c *CLIClient) exec(ctx context.Context, command string) error {
	out, err := c.run(ctx, command)
	if err != nil {
		return err
	}
	if commandRejected(out) {
		return fmt.Errorf("%w: %s", olt.ErrCommandRejected, firstMeaningfulLine(out))
	}
	return nil
}

func (c *CLIClient) enterPONInterface(ctx context.Context, ponIndex int) error {
	if _, err := c.run(ctx, "conf t"); err != nil {
		return err
	}
	if _, err := c.run(ctx, fmt.Sprintf("interface epon 0/%d", ponIndex)); err != nil {
		return err
	}
	return nil
}

// exitToPrivileged keluar dari mode config sampai prompt enable kembali.
func (c *CLIClient) exitToPrivileged(ctx context.Context) {
	for i := 0; i < 4; i++ {
		out, err := c.run(ctx, "exit")
		if err != nil {
			return
		}
		if isPrivilegedPrompt(lastLine(out)) {
			return
		}
	}
}

func isPrivilegedPrompt(prompt string) bool {
	prompt = strings.TrimSpace(prompt)
	return strings.HasSuffix(prompt, "#") && !strings.Contains(prompt, "(")
}

func firstMeaningfulLine(out string) string {
	for _, line := range strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n") {
		if line = strings.TrimSpace(line); line != "" && !strings.Contains(line, "EPON") {
			return line
		}
	}
	return strings.TrimSpace(out)
}

// Klasifikasi perintah untuk AI Router Ops (PLAN-AI-ROUTER-OPS v2).
//
// Hanya perintah READ yang boleh dieksekusi otomatis oleh pipeline AI;
// sisanya MUTATION dan butuh approval. Perintah tak dikenal = MUTATION
// (fail-closed). Reset/restore penuh di-hard-deny; reboot dan provisioning
// ONU (rename/reboot/delete ONU) adalah mutation biasa yang butuh approval.

// readPrefixes adalah prefix CLI HiOSO yang hanya membaca state.
var readPrefixes = []string{"show ", "display ", "ping ", "traceroute "}

// hardDenyPrefixes adalah perintah reset/restore penuh yang tidak boleh
// pernah diajukan AI.
var hardDenyPrefixes = []string{"reset", "restore", "factory-reset", "factory reset"}

// Classify menandai klasifikasi AI Router Ops untuk perintah CLI HiOSO.
func Classify(cmd command.Command) command.Classification {
	raw := strings.ToLower(strings.TrimSpace(cmd.Raw))
	if raw == "" {
		return command.Classification{Mode: command.ModeMutation}
	}
	for _, prefix := range hardDenyPrefixes {
		if raw == prefix || strings.HasPrefix(raw, prefix+" ") {
			return command.Classification{
				Mode:       command.ModeMutation,
				DenyReason: "hard deny for AI: " + prefix,
			}
		}
	}
	for _, prefix := range readPrefixes {
		if strings.HasPrefix(raw, prefix) {
			return command.Classification{Mode: command.ModeRead}
		}
	}
	return command.Classification{Mode: command.ModeMutation}
}

// Translate memetakan operation abstrak ke perintah CLI HiOSO.
func Translate(op command.Operation) (command.Command, error) {
	switch op {
	case command.OpGetStatus:
		return command.Command{Raw: "show version"}, nil
	case command.OpReboot:
		return command.Command{Raw: "reboot"}, nil
	default:
		return command.Command{}, fmt.Errorf("hioso: unsupported operation %q", op)
	}
}
