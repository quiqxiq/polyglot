package mikrotik

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// DedicatedQueueParams is the vendor-neutral dedicated queue params.
type DedicatedQueueParams = port.DedicatedQueueParams

// QueueGateway implements port.QueueGateway memakai command builder di
// queue.go dan eksekusi lewat policy-gated CommandExecutor.
type QueueGateway struct {
	exec port.CommandExecutor
}

// NewQueueGateway creates a QueueGateway bound to exec. exec must be the
// policy-gated executor (usecase/network.ExecuteCommand) — a bare
// driver.Execute here would silently bypass destructive-command approval.
func NewQueueGateway(exec port.CommandExecutor) *QueueGateway {
	return &QueueGateway{exec: exec}
}

var _ port.QueueGateway = (*QueueGateway)(nil)

// dedicatedQueueToROS builds /queue/simple/add from typed params.
func dedicatedQueueToROS(p DedicatedQueueParams) command.Command {
	args := map[string]string{
		"name":      p.Name,
		"target":    p.Target,
		"max-limit": p.MaxLimit,
		"limit-at":  p.LimitAt,
	}
	if p.Comment != "" {
		args["comment"] = p.Comment
	}
	if p.BurstLimit != "" {
		args["burst-limit"] = p.BurstLimit
	}
	if p.BurstThreshold != "" {
		args["burst-threshold"] = p.BurstThreshold
	}
	if p.BurstTime != "" {
		args["burst-time"] = p.BurstTime
	}
	return command.Command{Raw: "/queue/simple/add", Args: args}
}

// ListQueues implements port.QueueGateway.
func (g *QueueGateway) ListQueues(ctx context.Context, driver port.DeviceDriver, nameFilter string) ([]port.SimpleQueue, error) {
	res, err := g.exec(ctx, driver, NewPrintSimpleQueuesCommand(nameFilter))
	if err != nil {
		return nil, err
	}
	return ParseSimpleQueues(res), nil
}

// AddQueue implements port.QueueGateway.
func (g *QueueGateway) AddQueue(ctx context.Context, driver port.DeviceDriver, p DedicatedQueueParams) (command.Result, error) {
	return g.exec(ctx, driver, dedicatedQueueToROS(p))
}

// RemoveQueue implements port.QueueGateway.
func (g *QueueGateway) RemoveQueue(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return g.exec(ctx, driver, NewRemoveSimpleQueueCommand(rosID))
}

// SetQueueEnabled implements port.QueueGateway.
func (g *QueueGateway) SetQueueEnabled(ctx context.Context, driver port.DeviceDriver, rosID string, enabled bool) (command.Result, error) {
	val := "no"
	if !enabled {
		val = "yes"
	}
	cmd := command.Command{
		Raw:  "/queue/simple/set",
		Args: map[string]string{"numbers": rosID, "disabled": val},
	}
	return g.exec(ctx, driver, cmd)
}
