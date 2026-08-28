package queue

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// Gateway implements port.QueueGateway using the modular command builders and parsers.
type Gateway struct {
	exec port.CommandExecutor
}

// NewGateway creates a Gateway bound to exec.
func NewGateway(exec port.CommandExecutor) *Gateway {
	return &Gateway{exec: exec}
}

var _ port.QueueGateway = (*Gateway)(nil)

// ListQueues implements port.QueueGateway.
func (g *Gateway) ListQueues(ctx context.Context, driver port.DeviceDriver, nameFilter string) ([]port.SimpleQueue, error) {
	res, err := g.exec(ctx, driver, NewPrintSimpleQueuesCommand(nameFilter))
	if err != nil {
		return nil, err
	}
	return ParseSimpleQueues(res), nil
}

// AddQueue implements port.QueueGateway.
func (g *Gateway) AddQueue(ctx context.Context, driver port.DeviceDriver, p DedicatedQueueParams) (command.Result, error) {
	return g.exec(ctx, driver, DedicatedQueueToROS(p))
}

// RemoveQueue implements port.QueueGateway.
func (g *Gateway) RemoveQueue(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return g.exec(ctx, driver, NewRemoveSimpleQueueCommand(rosID))
}

// SetQueueEnabled implements port.QueueGateway.
func (g *Gateway) SetQueueEnabled(ctx context.Context, driver port.DeviceDriver, rosID string, enabled bool) (command.Result, error) {
	return g.exec(ctx, driver, NewSetSimpleQueueEnabledCommand(rosID, enabled))
}

