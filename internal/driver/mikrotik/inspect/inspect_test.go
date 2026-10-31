package inspect

import (
	"context"
	"testing"
	"time"

	"github.com/quixiq/polyglot/internal/domain/command"
)

type fakeExecutor struct {
	version string
	calls   int
	rows    map[string][]map[string]string
}

func (f *fakeExecutor) Execute(_ context.Context, cmd command.Command) (command.Result, error) {
	f.calls++
	if cmd.Raw == "/system/resource/print" {
		return command.Result{Rows: []map[string]string{{"version": f.version}}}, nil
	}
	req := cmd.Args["request"]
	return command.Result{Rows: f.rows[req]}, nil
}

func TestCapabilitiesV7(t *testing.T) {
	ex := &fakeExecutor{version: "7.16.2 (stable)"}
	c := NewClient(ex)
	caps, err := c.Capabilities(context.Background())
	if err != nil {
		t.Fatalf("capabilities: %v", err)
	}
	if !caps.InspectSupported || caps.AIMode != AIModeFull {
		t.Fatalf("expected full v7 capabilities, got %+v", caps)
	}
	if caps.Major != 7 || caps.Minor != 16 || caps.Patch != 2 {
		t.Fatalf("version parse wrong: %+v", caps)
	}
	// Cache: second call must not hit the device.
	before := ex.calls
	if _, err := c.Capabilities(context.Background()); err != nil {
		t.Fatal(err)
	}
	if ex.calls != before {
		t.Fatalf("capabilities not cached: calls %d → %d", before, ex.calls)
	}
}

func TestCapabilitiesV6Degraded(t *testing.T) {
	c := NewClient(&fakeExecutor{version: "6.49.10 (long-term)"})
	caps, err := c.Capabilities(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if caps.InspectSupported || caps.AIMode != AIModeDegraded || caps.Warning == "" {
		t.Fatalf("expected degraded v6, got %+v", caps)
	}
	res, err := c.Validate(context.Background(), "/ip/address/add address=1.1.1.1/24")
	if err != nil {
		t.Fatal(err)
	}
	if res.Supported || res.Valid {
		t.Fatalf("v6 validate must be unsupported/invalid, got %+v", res)
	}
}

func TestInspectCallsAndParsing(t *testing.T) {
	ex := &fakeExecutor{
		version: "7.16.2",
		rows: map[string][]map[string]string{
			"child":      {{"type": "dir", "name": "address", "node-type": "dir"}},
			"syntax":     {{"type": "arg", "symbol": "address", "symbol-type": "argument", "text": "address"}},
			"completion": {{"type": "completion", "completion": "address=", "offset": "0", "preference": "10", "show": "true", "text": "address="}},
			"highlight":  {{"type": "error", "text": "bogus", "start": "5", "end": "9"}},
		},
	}
	c := NewClient(ex)
	ctx := context.Background()

	children, err := c.Children(ctx, "ip")
	if err != nil || len(children) != 1 || children[0].Name != "address" {
		t.Fatalf("children: %v %+v", err, children)
	}
	syntax, err := c.Syntax(ctx, "ip,address,add")
	if err != nil || len(syntax) != 1 || syntax[0].Symbol != "address" {
		t.Fatalf("syntax: %v %+v", err, syntax)
	}
	completion, err := c.Completion(ctx, "/ip/addr/set ", "ip,address")
	if err != nil || len(completion) != 1 || !completion[0].Show || completion[0].Offset != 0 {
		t.Fatalf("completion: %v %+v", err, completion)
	}
	highlights, err := c.Highlight(ctx, "/ip/address/add bogus=1")
	if err != nil || len(highlights) != 1 || highlights[0].Start != 5 {
		t.Fatalf("highlight: %v %+v", err, highlights)
	}
	res, err := c.Validate(ctx, "/ip/address/add bogus=1")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Supported || res.Valid || len(res.Errors) != 1 {
		t.Fatalf("expected invalid validation, got %+v", res)
	}
}

func TestValidateRejectsInvalidInput(t *testing.T) {
	c := NewClient(&fakeExecutor{version: "7.16.2"})
	if _, err := c.Children(context.Background(), "../etc"); err == nil {
		t.Fatal("path traversal must be rejected")
	}
	if _, err := c.Children(context.Background(), ""); err == nil {
		t.Fatal("empty path must be rejected")
	}
	tooLong := make([]byte, maxInputBytes+1)
	for i := range tooLong {
		tooLong[i] = 'a'
	}
	if _, err := c.Highlight(context.Background(), string(tooLong)); err == nil {
		t.Fatal("oversized input must be rejected")
	}
}

func TestScriptUnsafeVersionSkipsSyntax(t *testing.T) {
	caps := BuildCapabilities("7.16.2", 7, 16, 2)
	if !caps.ScriptPathsUnsafe("ip/firewall/filter where") {
		t.Fatal("scripting path must be unsafe on 7.16.2")
	}
	fixed := BuildCapabilities("7.21.4", 7, 21, 4)
	if fixed.ScriptPathsUnsafe("ip/firewall/filter where") {
		t.Fatal("scripting path should be safe on 7.21.4")
	}
	if caps.ScriptPathsUnsafe("ip/address") {
		t.Fatal("normal path should be safe")
	}
}

func TestCapabilitiesExpireCache(t *testing.T) {
	ex := &fakeExecutor{version: "7.16.2"}
	c := NewClient(ex)
	c.cacheTTL = time.Nanosecond
	if _, err := c.Capabilities(context.Background()); err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Nanosecond)
	if _, err := c.Capabilities(context.Background()); err != nil {
		t.Fatal(err)
	}
	if ex.calls < 2 {
		t.Fatalf("expired cache should re-read, calls=%d", ex.calls)
	}
}
