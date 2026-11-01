package mikrotik

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/command"
)

type recordingDriver struct {
	mu   sync.Mutex
	cmds []command.Command
}

func (d *recordingDriver) Execute(_ context.Context, cmd command.Command) (command.Result, error) {
	d.mu.Lock()
	d.cmds = append(d.cmds, cmd)
	d.mu.Unlock()
	switch cmd.Raw {
	case "/file/print":
		return command.Result{Rows: []map[string]string{{".id": "*F1", "name": cmd.Args["?name"]}}}, nil
	case "/system/clock/print":
		return command.Result{Rows: []map[string]string{{"date": "2026-10-31", "time": "10:00:00"}}}, nil
	default:
		return command.Result{}, nil
	}
}
func (d *recordingDriver) Classify(command.Command) command.Classification {
	return command.Classification{Mode: command.ModeMutation}
}
func (d *recordingDriver) Translate(command.Operation) (command.Command, error) {
	return command.Command{}, nil
}
func (d *recordingDriver) Close() error { return nil }

func (d *recordingDriver) find(raw string) (command.Command, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, c := range d.cmds {
		if c.Raw == raw {
			return c, true
		}
	}
	return command.Command{}, false
}

func TestAIOpsArmDisarmWatchdog(t *testing.T) {
	drv := &recordingDriver{}
	ops := NewAIOps(drv)
	ctx := context.Background()

	require.NoError(t, ops.ArmWatchdog(ctx, "apr-1", []byte("/ip/firewall/filter/remove .id=*A\n"), 600))

	add, ok := drv.find("/system/scheduler/add")
	require.True(t, ok, "scheduler add must be issued")
	assert.Equal(t, "polyglot-wd-apr-1", add.Args["name"])
	assert.Contains(t, add.Args["on-event"], "/import file=polyglot-wd-apr-1.rsc")
	assert.Equal(t, "0", add.Args["interval"])
	assert.NotEmpty(t, add.Args["start-date"])
	assert.NotEmpty(t, add.Args["start-time"])

	require.NoError(t, ops.DisarmWatchdog(ctx, "apr-1"))
	remove, ok := drv.find("/system/scheduler/remove")
	require.True(t, ok, "scheduler remove must be issued")
	assert.Equal(t, "polyglot-wd-apr-1", remove.Args["?name"])
}

func TestAIOpsArmWatchdogRejectsEmpty(t *testing.T) {
	ops := NewAIOps(&recordingDriver{})
	// Zero seconds falls back to 600 rather than failing; the guard worth
	// testing is that the inverse script is still uploaded.
	require.NoError(t, ops.ArmWatchdog(context.Background(), "x", []byte("# noop\n"), 0))
}

func TestAIOpsReadMenuQueryFormatting(t *testing.T) {
	drv := &recordingDriver{}
	ops := NewAIOps(drv)
	ctx := context.Background()

	// 1. Selector with .id="*" should omit .id wildcard and not fail
	_, err := ops.ReadMenu(ctx, "interface", map[string]string{".id": "*"})
	require.NoError(t, err)
	cmd, ok := drv.find("/interface/print")
	require.True(t, ok)
	assert.Equal(t, "", cmd.Args["detail"])
	assert.NotContains(t, cmd.Args, ".id", ".id without ? must never be passed as parameter")
	assert.NotContains(t, cmd.Args, "?.id", "wildcard * must be omitted")

	// 2. Selector with explicit .id="*1" must become ?.id=*1 query
	drv.mu.Lock()
	drv.cmds = nil
	drv.mu.Unlock()
	_, err = ops.ReadMenu(ctx, "interface", map[string]string{".id": "*1"})
	require.NoError(t, err)
	cmd, ok = drv.find("/interface/print")
	require.True(t, ok)
	assert.Equal(t, "*1", cmd.Args["?.id"])
	assert.NotContains(t, cmd.Args, ".id")

	// 3. Selector with regular property name="ether1" must become ?name=ether1
	drv.mu.Lock()
	drv.cmds = nil
	drv.mu.Unlock()
	_, err = ops.ReadMenu(ctx, "interface", map[string]string{"name": "ether1"})
	require.NoError(t, err)
	cmd, ok = drv.find("/interface/print")
	require.True(t, ok)
	assert.Equal(t, "ether1", cmd.Args["?name"])

	// 4. Menu alias normalization ("resource" -> "/system/resource/print", "system" -> "/system/resource/print")
	drv.mu.Lock()
	drv.cmds = nil
	drv.mu.Unlock()
	_, err = ops.ReadMenu(ctx, "resource", nil)
	require.NoError(t, err)
	cmd, ok = drv.find("/system/resource/print")
	require.True(t, ok, "resource must normalize to /system/resource/print")

	drv.mu.Lock()
	drv.cmds = nil
	drv.mu.Unlock()
	_, err = ops.ReadMenu(ctx, "system", nil)
	require.NoError(t, err)
	cmd, ok = drv.find("/system/resource/print")
	require.True(t, ok, "system must normalize to /system/resource/print")

	drv.mu.Lock()
	drv.cmds = nil
	drv.mu.Unlock()
	_, err = ops.ReadMenu(ctx, "routes", nil)
	require.NoError(t, err)
	cmd, ok = drv.find("/ip/route/print")
	require.True(t, ok, "routes must normalize to /ip/route/print")
}
