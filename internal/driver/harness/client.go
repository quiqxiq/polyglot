// Package harness runs one-shot DeepSeek harness turns as a child process.
//
// The harness never owns a WhatsApp socket or a long-lived session here. Each
// call starts the installed dsh CLI with the read-only customer-service overlay,
// writes the prompt to stdin, and reads the terminal JSON event from stdout.
package harness

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultMaxLineBytes   = 8 << 20
	defaultMaxOutputBytes = 64 << 20
	maxErrorMessageRunes  = 500
)

var (
	// ErrUnavailable reports that the harness could not start or finish a run.
	ErrUnavailable = errors.New("harness unavailable")
	// ErrTimeout reports that the harness exceeded its configured run timeout.
	ErrTimeout = errors.New("harness timed out")
	// ErrNoReply reports that a successful run produced no final assistant text.
	ErrNoReply = errors.New("harness returned no reply")
)

// Config describes one installed harness profile invocation.
type Config struct {
	// Bin is the dsh executable name or path.
	Bin string
	// PatchPath is the customer-service overlay applied after the profile layer.
	PatchPath string
	// WorkDir is the child process working directory. Empty inherits the parent.
	WorkDir string
	// MCPURL is exported as POLYGLOT_MCP_URL when non-empty.
	MCPURL string
	// MCPReadToken is exported as POLYGLOT_MCP_READ_TOKEN when non-empty.
	MCPReadToken string
	// Timeout bounds one complete run.
	Timeout time.Duration
	// MaxConcurrent bounds simultaneously running harness processes.
	MaxConcurrent int
}

// Client executes headless harness turns.
type Client struct {
	bin          string
	patchPath    string
	workDir      string
	dshHome      string
	mcpURL       string
	mcpReadToken string
	timeout      time.Duration
	sem          chan struct{}
}

// New validates the harness installation and returns a bounded client.
func New(cfg Config) (*Client, error) {
	bin := strings.TrimSpace(cfg.Bin)
	if bin == "" {
		return nil, fmt.Errorf("%w: HARNESS_DSH_BIN is empty", ErrUnavailable)
	}
	resolvedBin, err := exec.LookPath(bin)
	if err != nil {
		return nil, fmt.Errorf("%w: executable %q not found: %w", ErrUnavailable, bin, err)
	}

	workDir := strings.TrimSpace(cfg.WorkDir)
	dshHome := ""
	if workDir == "" {
		// Run from the harness home, not the Polyglot project root: dsh loads
		// the invoking directory's .env, which would reintroduce the server
		// secrets scrubbed from the child environment.
		dshHome, err = resolveHarnessHome()
		if err != nil {
			return nil, fmt.Errorf("%w: resolve harness home: %w", ErrUnavailable, err)
		}
		workDir = dshHome
		if err := os.MkdirAll(workDir, 0o700); err != nil {
			return nil, fmt.Errorf("%w: create harness home %q: %w", ErrUnavailable, workDir, err)
		}
	}
	workDir, err = filepath.Abs(workDir)
	if err != nil {
		return nil, fmt.Errorf("%w: resolve harness working directory: %w", ErrUnavailable, err)
	}
	if info, err := os.Stat(workDir); err != nil || !info.IsDir() {
		return nil, fmt.Errorf("%w: harness working directory %q is not a directory", ErrUnavailable, workDir)
	}

	patchPath := strings.TrimSpace(cfg.PatchPath)
	if patchPath == "" {
		return nil, fmt.Errorf("%w: HARNESS_CS_PATCH is empty", ErrUnavailable)
	}
	if !filepath.IsAbs(patchPath) {
		base, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("%w: resolve working directory: %w", ErrUnavailable, err)
		}
		patchPath = filepath.Join(base, patchPath)
	}
	if _, err := os.Stat(patchPath); err != nil {
		return nil, fmt.Errorf("%w: customer-service patch %q: %w", ErrUnavailable, patchPath, err)
	}

	if cfg.Timeout <= 0 {
		return nil, fmt.Errorf("%w: HARNESS_TIMEOUT_SECONDS must be positive", ErrUnavailable)
	}
	if cfg.MaxConcurrent <= 0 {
		return nil, fmt.Errorf("%w: HARNESS_MAX_CONCURRENT must be positive", ErrUnavailable)
	}

	return &Client{
		bin:          resolvedBin,
		patchPath:    patchPath,
		workDir:      workDir,
		dshHome:      dshHome,
		mcpURL:       strings.TrimSpace(cfg.MCPURL),
		mcpReadToken: strings.TrimSpace(cfg.MCPReadToken),
		timeout:      cfg.Timeout,
		sem:          make(chan struct{}, cfg.MaxConcurrent),
	}, nil
}

// resolveHarnessHome returns the dsh home used when no explicit working
// directory is configured.
func resolveHarnessHome() (string, error) {
	if configured := strings.TrimSpace(os.Getenv("DSH_HOME")); configured != "" {
		return filepath.Abs(configured)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".dsh"), nil
}

// Ask runs one headless turn and returns its committed final assistant text.
func (c *Client) Ask(ctx context.Context, prompt string) (string, error) {
	if strings.TrimSpace(prompt) == "" {
		return "", fmt.Errorf("%w: empty prompt", ErrNoReply)
	}

	select {
	case c.sem <- struct{}{}:
		defer func() { <-c.sem }()
	case <-ctx.Done():
		return "", ctx.Err()
	}

	runCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, c.bin,
		"--profile", "headless",
		"--patch", c.patchPath,
		"--approval-policy", "never",
		"--json", "-",
	)
	cmd.Dir = c.workDir
	cmd.Env = c.environment()
	cmd.Stdin = strings.NewReader(prompt)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("%w: open stdout pipe: %w", ErrUnavailable, err)
	}
	var stderr boundedBuffer
	stderr.limit = 4 << 10
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("%w: start process: %w", ErrUnavailable, err)
	}

	final, streamErr := readFinal(stdout)
	if streamErr != nil {
		// Stop the writer before Wait so a broken stream cannot block on a
		// full pipe while the scanner has already given up.
		cancel()
	}
	waitErr := cmd.Wait()

	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	if runCtx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("%w after %s", ErrTimeout, c.timeout)
	}
	if streamErr != nil {
		return "", streamErr
	}
	if waitErr != nil {
		return "", fmt.Errorf("%w: process exited with %w", ErrUnavailable, waitErr)
	}
	if final == "" {
		return "", ErrNoReply
	}
	return final, nil
}

// harnessScrubbedEnv lists server credentials the one-shot harness must never
// inherit. The child only needs its own DeepSeek credentials and the read-only
// MCP token supplied below.
var harnessScrubbedEnv = []string{
	"JWT_SECRET",
	"ENCRYPTION_KEY",
	"DATABASE_URL",
	"REDIS_URL",
	"POLYGLOT_MCP_TOKEN",
	"POLYGLOT_WS_TOKEN",
	"HARNESS_WEB_TOKEN",
	"HARNESS_PROXY_SECRET",
}

func (c *Client) environment() []string {
	env := os.Environ()
	for _, key := range harnessScrubbedEnv {
		env = unsetEnv(env, key)
	}
	env = unsetEnv(env, "POLYGLOT_MCP_URL")
	env = unsetEnv(env, "POLYGLOT_MCP_READ_TOKEN")
	if c.dshHome != "" {
		// Pin the home the client resolved so dsh cannot fall back to a
		// different profile or session-log root.
		env = unsetEnv(env, "DSH_HOME")
		env = append(env, "DSH_HOME="+c.dshHome)
	}
	if c.mcpURL != "" {
		env = append(env, "POLYGLOT_MCP_URL="+c.mcpURL)
	}
	if c.mcpReadToken != "" {
		env = append(env, "POLYGLOT_MCP_READ_TOKEN="+c.mcpReadToken)
	}
	return env
}

// unsetEnv removes every occurrence of key from the child environment.
func unsetEnv(env []string, key string) []string {
	prefix := key + "="
	kept := make([]string, 0, len(env))
	for _, entry := range env {
		if !strings.HasPrefix(entry, prefix) {
			kept = append(kept, entry)
		}
	}
	return kept
}

type jsonEvent struct {
	Type    string `json:"type"`
	Text    string `json:"text"`
	Message string `json:"message"`
	Reason  *struct {
		Kind string `json:"kind"`
	} `json:"reason"`
}

// readFinal scans the newline-delimited run projection and returns the last
// completed final text. A malformed line is ignored because diagnostics belong
// on stderr; an error event or a non-completed final is reported.
func readFinal(r io.Reader) (string, error) {
	limited := &io.LimitedReader{R: r, N: defaultMaxOutputBytes + 1}
	scanner := bufio.NewScanner(limited)
	scanner.Buffer(make([]byte, 64<<10), defaultMaxLineBytes)

	var final string
	var finalErr error
	var streamErr error

	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var evt jsonEvent
		if err := json.Unmarshal(line, &evt); err != nil {
			continue
		}
		switch evt.Type {
		case "final":
			final = evt.Text
			if evt.Reason != nil && evt.Reason.Kind != "" && evt.Reason.Kind != "completed" {
				finalErr = fmt.Errorf("%w: run ended with %s", ErrUnavailable, evt.Reason.Kind)
			} else {
				finalErr = nil
			}
		case "error":
			message := strings.TrimSpace(evt.Message)
			if message == "" {
				message = "harness reported an error"
			}
			streamErr = fmt.Errorf("%w: %s", ErrUnavailable, truncateRunes(message, maxErrorMessageRunes))
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("%w: read event stream: %w", ErrUnavailable, err)
	}
	if limited.N <= 0 {
		return "", fmt.Errorf("%w: event stream exceeded %d bytes", ErrUnavailable, defaultMaxOutputBytes)
	}
	if finalErr != nil {
		return "", finalErr
	}
	if streamErr != nil && final == "" {
		return "", streamErr
	}
	return final, nil
}

type boundedBuffer struct {
	buf   bytes.Buffer
	limit int
}

func (b *boundedBuffer) Write(p []byte) (int, error) {
	n := len(p)
	if b.limit <= 0 {
		return n, nil
	}
	if room := b.limit - b.buf.Len(); room > 0 {
		if len(p) > room {
			p = p[:room]
		}
		b.buf.Write(p)
	}
	return n, nil
}

func truncateRunes(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit]) + "…"
}
