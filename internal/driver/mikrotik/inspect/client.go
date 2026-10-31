package inspect

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// Executor is the subset of a device driver that inspect needs.
type Executor interface {
	Execute(ctx context.Context, cmd command.Command) (command.Result, error)
}

// validPath matches RouterOS menu paths. Anything else is rejected before it
// reaches the device (defense against injection through the API path word).
var validPath = regexp.MustCompile(`^[a-zA-Z0-9_./,-]+$`)

const (
	maxInputBytes = 8 * 1024
	// defaultTimeout bounds one /console/inspect round-trip.
	defaultTimeout = 5 * time.Second
	// defaultCacheTTL is the capability cache lifetime per device.
	defaultCacheTTL = 10 * time.Minute
)

// Client wraps /console/inspect for one device.
type Client struct {
	exec     Executor
	timeout  time.Duration
	cacheTTL time.Duration

	mu     sync.Mutex
	caps   *Capabilities
	capsAt time.Time
}

// NewClient builds an inspect client. exec is usually the MikroTik driver.
func NewClient(exec Executor) *Client {
	return &Client{exec: exec, timeout: defaultTimeout, cacheTTL: defaultCacheTTL}
}

// WithTimeout overrides the per-call timeout.
func (c *Client) WithTimeout(d time.Duration) *Client {
	if d > 0 {
		c.timeout = d
	}
	return c
}

// Children calls request=child for a menu path.
func (c *Client) Children(ctx context.Context, path string) ([]Node, error) {
	if err := validatePath(path); err != nil {
		return nil, err
	}
	res, err := c.run(ctx, NewChildCommand(path))
	if err != nil {
		return nil, err
	}
	return ParseChild(res), nil
}

// Syntax calls request=syntax for a path (e.g. "ip,address,add").
func (c *Client) Syntax(ctx context.Context, path string) ([]Syntax, error) {
	if err := validatePath(path); err != nil {
		return nil, err
	}
	caps, err := c.Capabilities(ctx)
	if err != nil {
		return nil, err
	}
	if !caps.InspectSupported || caps.ScriptPathsUnsafe(path) {
		return nil, fmt.Errorf("inspect: syntax unsupported on %s", caps.Version)
	}
	res, err := c.run(ctx, NewSyntaxCommand(path))
	if err != nil {
		return nil, err
	}
	return ParseSyntax(res), nil
}

// Completion calls request=completion.
func (c *Client) Completion(ctx context.Context, input, path string) ([]Completion, error) {
	if err := validateInput(input); err != nil {
		return nil, err
	}
	if path != "" {
		if err := validatePath(path); err != nil {
			return nil, err
		}
	}
	caps, err := c.Capabilities(ctx)
	if err != nil {
		return nil, err
	}
	if !caps.InspectSupported || caps.ScriptPathsUnsafe(path) {
		return nil, fmt.Errorf("inspect: completion unsupported on %s", caps.Version)
	}
	res, err := c.run(ctx, NewCompletionCommand(input, path))
	if err != nil {
		return nil, err
	}
	return ParseCompletion(res), nil
}

// Highlight calls request=highlight.
func (c *Client) Highlight(ctx context.Context, input string) ([]Highlight, error) {
	if err := validateInput(input); err != nil {
		return nil, err
	}
	res, err := c.run(ctx, NewHighlightCommand(input))
	if err != nil {
		return nil, err
	}
	return ParseHighlight(res), nil
}

// Validate checks a script's syntax via request=highlight. On RouterOS v6 the
// result is Supported=false (degraded), never an error/crash.
func (c *Client) Validate(ctx context.Context, script string) (ValidationResult, error) {
	caps, err := c.Capabilities(ctx)
	if err != nil {
		return ValidationResult{}, err
	}
	if !caps.InspectSupported {
		return ValidationResult{
			Supported: false,
			Valid:     false,
			Warnings:  []string{caps.Warning},
		}, nil
	}
	highlights, err := c.Highlight(ctx, script)
	if err != nil {
		return ValidationResult{}, err
	}
	return BuildValidation(highlights), nil
}

// Capabilities returns the cached device capability profile.
func (c *Client) Capabilities(ctx context.Context) (Capabilities, error) {
	c.mu.Lock()
	if c.caps != nil && time.Since(c.capsAt) < c.cacheTTL {
		out := *c.caps
		c.mu.Unlock()
		return out, nil
	}
	c.mu.Unlock()

	res, err := c.exec.Execute(ctx, command.Command{
		Raw:  "/system/resource/print",
		Args: map[string]string{".proplist": "version"},
	})
	if err != nil {
		return Capabilities{}, fmt.Errorf("inspect: read router version: %w", err)
	}
	version, major, minor, patch, err := ParseCapabilities(res)
	if err != nil {
		return Capabilities{}, err
	}
	caps := BuildCapabilities(version, major, minor, patch)

	c.mu.Lock()
	c.caps = &caps
	c.capsAt = time.Now()
	c.mu.Unlock()
	return caps, nil
}

func (c *Client) run(ctx context.Context, cmd command.Command) (command.Result, error) {
	runCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return c.exec.Execute(runCtx, cmd)
}

func validatePath(path string) error {
	p := strings.TrimSpace(path)
	if p == "" {
		return fmt.Errorf("inspect: path is required")
	}
	if strings.Contains(p, "..") || !validPath.MatchString(p) {
		return fmt.Errorf("inspect: invalid path %q", path)
	}
	return nil
}

func validateInput(input string) error {
	if strings.TrimSpace(input) == "" {
		return fmt.Errorf("inspect: input is required")
	}
	if len(input) > maxInputBytes {
		return fmt.Errorf("inspect: input exceeds %d bytes", maxInputBytes)
	}
	return nil
}
