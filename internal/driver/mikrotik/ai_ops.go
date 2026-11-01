package mikrotik

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/restorepoint"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/inspect"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/system"
	"github.com/quixiq/polyglot/internal/port"
)

// AIOps adapts a MikroTik driver to the vendor-neutral port.RouterOps surface
// consumed by the AI Router Ops usecases (restore point, diff, execution).
type AIOps struct {
	exec    port.DeviceDriver
	inspect *inspect.Client
}

var _ port.RouterOps = (*AIOps)(nil)

// NewAIOps wraps a connected device driver (usually *Driver).
func NewAIOps(exec port.DeviceDriver) *AIOps {
	return &AIOps{exec: exec, inspect: inspect.NewClient(exec)}
}

// Capabilities reports router version and AI support mode.
func (a *AIOps) Capabilities(ctx context.Context) (port.DeviceCapabilities, error) {
	caps, err := a.inspect.Capabilities(ctx)
	if err != nil {
		return port.DeviceCapabilities{}, err
	}
	return port.DeviceCapabilities{
		Version:          caps.Version,
		Major:            caps.Major,
		InspectSupported: caps.InspectSupported,
		AIMode:           string(caps.AIMode),
		Warning:          caps.Warning,
	}, nil
}

// ValidateScript pre-validates a command via /console/inspect highlight.
func (a *AIOps) ValidateScript(ctx context.Context, script string) (port.ScriptValidation, error) {
	res, err := a.inspect.Validate(ctx, script)
	if err != nil {
		return port.ScriptValidation{}, err
	}
	out := port.ScriptValidation{
		Supported: res.Supported,
		Valid:     res.Valid,
		Warnings:  res.Warnings,
	}
	for _, e := range res.Errors {
		out.Errors = append(out.Errors, fmt.Sprintf("%s at %d-%d", e.Text, e.Start, e.End))
	}
	return out, nil
}

// ReadMenu reads rows from a menu (print detail) with a unique selector.
func (a *AIOps) ReadMenu(ctx context.Context, menu string, selector map[string]string) ([]map[string]string, error) {
	normMenu := NormalizeMenuPath(menu)
	args := map[string]string{"detail": ""}
	for k, v := range selector {
		val := strings.TrimSpace(v)
		if val == "" || val == "*" {
			continue
		}
		key := strings.TrimSpace(k)
		if !strings.HasPrefix(key, "?") {
			key = "?" + key
		}
		args[key] = val
	}
	res, err := a.exec.Execute(ctx, command.Command{Raw: commandRaw(normMenu, "print"), Args: args})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no such command or directory") {
			return nil, fmt.Errorf("%w (petunjuk menu RouterOS: 'system/resource', 'system/health', 'system/identity', 'interface', 'ip/address', 'ip/route', 'ip/firewall/filter', 'ip/dhcp-server/lease')", err)
		}
		return nil, err
	}
	return res.Rows, nil
}

// CreateArtifacts creates scoped exports, a deduplicated binary backup, and
// the inverse script artifact. Router temp files are removed best-effort.
func (a *AIOps) CreateArtifacts(ctx context.Context, restorePointID string, scopes []string, inverseScript []byte, maxBytes int64) ([]port.RawArtifact, error) {
	var out []port.RawArtifact
	cleanup := make([]string, 0, len(scopes)+1)
	defer func() {
		for _, name := range cleanup {
			a.RemoveFile(ctx, name)
		}
	}()

	for i, scope := range scopes {
		slug := sanitizeSlug(scope)
		if slug == "" {
			slug = fmt.Sprintf("scope-%d", i)
		}
		name := fmt.Sprintf("polyglot-rp-%s-%s", sanitizeSlug(restorePointID), slug)
		if _, err := a.exec.Execute(ctx, system.NewScopedExportCommand(scope, name)); err != nil {
			return nil, fmt.Errorf("export scope %q: %w", scope, err)
		}
		cleanup = append(cleanup, name)
		if err := a.waitSettled(ctx, name, 15*time.Second); err != nil {
			return nil, err
		}
		data, size, err := a.readFile(ctx, name, maxBytes)
		if err != nil {
			return nil, err
		}
		out = append(out, port.RawArtifact{
			Kind: restorepoint.ArtifactScopedExport, Scope: []string{scope},
			Name: name + ".rsc", Data: io.NopCloser(bytes.NewReader(data)), Size: size,
		})
	}

	binName := fmt.Sprintf("polyglot-rp-%s-bin", sanitizeSlug(restorePointID))
	if _, err := a.exec.Execute(ctx, system.NewBackupSaveCommand(system.BackupSaveConfig{Name: binName, DontEncrypt: true})); err != nil {
		return nil, fmt.Errorf("binary backup: %w", err)
	}
	cleanup = append(cleanup, binName)
	if err := a.waitSettled(ctx, binName, 30*time.Second); err != nil {
		return nil, err
	}
	binData, binSize, err := a.readFile(ctx, binName, maxBytes)
	if err != nil {
		return nil, err
	}
	out = append(out, port.RawArtifact{
		Kind: restorepoint.ArtifactBinaryBackup, Name: binName + ".backup",
		Data: io.NopCloser(bytes.NewReader(binData)), Size: binSize,
	})

	if len(inverseScript) > 0 {
		out = append(out, port.RawArtifact{
			Kind: restorepoint.ArtifactInverseScript, Name: "inverse.rsc",
			Data: io.NopCloser(bytes.NewReader(inverseScript)), Size: int64(len(inverseScript)),
		})
	}
	return out, nil
}

// ExecuteRaw executes one vendor-native command through the driver.
func (a *AIOps) ExecuteRaw(ctx context.Context, raw string, args map[string]string) (command.Result, error) {
	return a.exec.Execute(ctx, command.Command{Raw: raw, Args: args})
}

// UploadFile writes a small script/backup to the router. RouterOS /file/set
// accepts one contents chunk (~60KB); larger files are rejected in v1.
func (a *AIOps) UploadFile(ctx context.Context, name string, data []byte) error {
	const maxUpload = 60000
	if len(data) > maxUpload {
		return fmt.Errorf("upload %q: %d bytes exceeds %d-byte v1 limit", name, len(data), maxUpload)
	}
	if _, err := a.exec.Execute(ctx, system.NewAddFileCommand(name)); err != nil {
		return fmt.Errorf("add file %q: %w", name, err)
	}
	id, err := a.fileID(ctx, name)
	if err != nil {
		return err
	}
	_, err = a.exec.Execute(ctx, system.NewSetFileCommand(id, string(data)))
	return err
}

// ImportScript applies an uploaded script via /import.
func (a *AIOps) ImportScript(ctx context.Context, name string) error {
	_, err := a.exec.Execute(ctx, command.Command{Raw: "/import", Args: map[string]string{"file": name}})
	return err
}

// LoadBinary restores an uploaded binary backup and reboots the router.
func (a *AIOps) LoadBinary(ctx context.Context, name string) error {
	if _, err := a.exec.Execute(ctx, system.NewBackupLoadCommand(name, "")); err != nil {
		return err
	}
	_, err := a.exec.Execute(ctx, command.Command{Raw: "/system/reboot"})
	return err
}

// RemoveFile deletes a temporary router file (best effort).
func (a *AIOps) RemoveFile(ctx context.Context, name string) {
	id, err := a.fileID(ctx, name)
	if err != nil || id == "" {
		return
	}
	_, _ = a.exec.Execute(ctx, system.NewRemoveFileCommand(id))
}

// PostVerify re-reads affected menus after a mutation.
func (a *AIOps) PostVerify(ctx context.Context, menus []string) (map[string][]map[string]string, error) {
	out := make(map[string][]map[string]string, len(menus))
	for _, menu := range menus {
		rows, err := a.ReadMenu(ctx, menu, nil)
		if err != nil {
			return out, fmt.Errorf("post-verify %s: %w", menu, err)
		}
		out[menu] = rows
	}
	return out, nil
}

// ResourceInfo returns version and free disk space for preflight.
func (a *AIOps) ResourceInfo(ctx context.Context) (string, int64, error) {
	res, err := a.exec.Execute(ctx, system.NewResourcePrintCommand())
	if err != nil {
		return "", 0, err
	}
	if len(res.Rows) == 0 {
		return "", 0, fmt.Errorf("empty resource info")
	}
	row := res.Rows[0]
	return row["version"], parseRouterSize(row["free-hdd-space"]), nil
}

// RouterNow returns the router clock (falls back to server UTC on parse error).
func (a *AIOps) RouterNow(ctx context.Context) (time.Time, error) {
	res, err := a.exec.Execute(ctx, system.NewClockPrintCommand())
	if err != nil {
		return time.Time{}, err
	}
	if len(res.Rows) > 0 {
		if t, ok := parseRouterClock(res.Rows[0]["date"], res.Rows[0]["time"]); ok {
			return t, nil
		}
	}
	return time.Now().UTC(), nil
}

func (a *AIOps) fileID(ctx context.Context, name string) (string, error) {
	res, err := a.exec.Execute(ctx, system.NewPrintFileCommand(name))
	if err != nil {
		return "", err
	}
	for _, row := range res.Rows {
		if row["name"] == name {
			return row[".id"], nil
		}
	}
	return "", fmt.Errorf("file %q not found", name)
}

func (a *AIOps) waitSettled(ctx context.Context, name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	last := int64(-1)
	stable := 0
	for time.Now().Before(deadline) {
		res, err := a.exec.Execute(ctx, system.NewPrintFileCommand(name))
		if err == nil && len(res.Rows) > 0 {
			size := system.ParseFileSize(res)
			if size > 0 {
				if size == last {
					stable++
					if stable >= 2 {
						return nil
					}
				} else {
					stable = 0
				}
				last = size
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(250 * time.Millisecond):
		}
	}
	return fmt.Errorf("file %q did not settle within %s", name, timeout)
}

func (a *AIOps) readFile(ctx context.Context, name string, maxBytes int64) ([]byte, int64, error) {
	var buf bytes.Buffer
	const chunkSize = 60000
	offset := 0
	var total int64
	for {
		res, err := a.exec.Execute(ctx, system.NewReadFileChunkCommand(name, offset, chunkSize))
		if err != nil {
			return nil, 0, fmt.Errorf("read file %q at %d: %w", name, offset, err)
		}
		chunk, ok := system.ParseReadChunk(res)
		if !ok || len(chunk) == 0 {
			break
		}
		buf.Write(chunk)
		total += int64(len(chunk))
		if maxBytes > 0 && total > maxBytes {
			return nil, 0, restorepoint.ErrSizeLimit
		}
		offset += len(chunk)
		if len(chunk) < chunkSize {
			break
		}
	}
	return buf.Bytes(), total, nil
}

func parseRouterSize(raw string) int64 {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0
	}
	units := map[string]int64{"kib": 1024, "mib": 1024 * 1024, "gib": 1024 * 1024 * 1024, "b": 1}
	lower := strings.ToLower(s)
	mult := int64(1)
	for suffix, m := range units {
		if strings.HasSuffix(lower, suffix) {
			mult = m
			s = strings.TrimSpace(s[:len(s)-len(suffix)])
			break
		}
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return int64(f * float64(mult))
}

func parseRouterClock(date, clock string) (time.Time, bool) {
	d := strings.TrimSpace(date)
	t := strings.TrimSpace(clock)
	if d == "" || t == "" {
		return time.Time{}, false
	}
	// RouterOS may return date as YYYY-MM-DD or DD-MMM-YYYY.
	for _, layout := range []string{"2006-01-02 15:04:05", "02-Jan-2006 15:04:05"} {
		if parsed, err := time.ParseInLocation(layout, d+" "+t, time.Local); err == nil {
			return parsed.UTC(), true
		}
	}
	return time.Time{}, false
}

// InspectChildren implements port.RouterInspector.
func (a *AIOps) InspectChildren(ctx context.Context, path string) ([]map[string]string, error) {
	nodes, err := a.inspect.Children(ctx, path)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]string, len(nodes))
	for i, n := range nodes {
		out[i] = map[string]string{"type": n.Type, "name": n.Name, "node-type": n.NodeType}
	}
	return out, nil
}

// InspectSyntax implements port.RouterInspector.
func (a *AIOps) InspectSyntax(ctx context.Context, path string) ([]map[string]string, error) {
	nodes, err := a.inspect.Syntax(ctx, path)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]string, len(nodes))
	for i, n := range nodes {
		out[i] = map[string]string{"type": n.Type, "symbol": n.Symbol, "symbol-type": n.SymbolType, "nested": n.Nested, "text": n.Text}
	}
	return out, nil
}

// InspectCompletion implements port.RouterInspector.
func (a *AIOps) InspectCompletion(ctx context.Context, input, path string) ([]map[string]string, error) {
	items, err := a.inspect.Completion(ctx, input, path)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]string, len(items))
	for i, n := range items {
		out[i] = map[string]string{"type": n.Type, "completion": n.Completion, "style": n.Style, "text": n.Text}
	}
	return out, nil
}

// InspectValidate implements port.RouterInspector.
func (a *AIOps) InspectValidate(ctx context.Context, script string) (port.ScriptValidation, error) {
	return a.ValidateScript(ctx, script)
}

var _ port.RouterInspector = (*AIOps)(nil)

// ArmWatchdog installs a RouterOS scheduler that imports the inverse script
// after `seconds` unless disarmed. It uses the router clock, not the server
// clock, and records the script as a temporary file.
func (a *AIOps) ArmWatchdog(ctx context.Context, approvalID string, inverseScript []byte, seconds int) error {
	if seconds <= 0 {
		seconds = 600
	}
	name := "polyglot-wd-" + sanitizeSlug(approvalID)
	script := name + ".rsc"
	if err := a.UploadFile(ctx, script, inverseScript); err != nil {
		return err
	}
	now, err := a.RouterNow(ctx)
	if err != nil {
		return err
	}
	start := now.Add(time.Duration(seconds) * time.Second)
	date := strings.ToLower(start.Format("Jan/02/2006"))
	clock := start.Format("15:04:05")
	_, err = a.exec.Execute(ctx, command.Command{Raw: "/system/scheduler/add", Args: map[string]string{
		"name":       name,
		"start-date": date,
		"start-time": clock,
		"interval":   "0",
		"on-event":   "/import file=" + script,
		"policy":     "ftp,reboot,read,write,policy,test,password,sniff,sensitive",
	}})
	if err != nil {
		a.RemoveFile(ctx, script)
		return err
	}
	return nil
}

// DisarmWatchdog removes the scheduler and its temporary script.
func (a *AIOps) DisarmWatchdog(ctx context.Context, approvalID string) error {
	name := "polyglot-wd-" + sanitizeSlug(approvalID)
	_, err := a.exec.Execute(ctx, command.Command{Raw: "/system/scheduler/remove", Args: map[string]string{"?name": name}})
	a.RemoveFile(ctx, name+".rsc")
	return err
}

var _ port.RollbackWatchdog = (*AIOps)(nil)
