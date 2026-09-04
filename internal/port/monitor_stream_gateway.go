package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// QueueStreamParams defines filter and duration parameters for streaming queue statistics.
type QueueStreamParams struct {
	NameFilter   string
	ParentFilter string
	ParentsOnly  bool
	Interval     string
}

// MonitorStreamGateway provides streaming operations and event parsing for device monitoring.
type MonitorStreamGateway interface {
	StreamTraffic(ctx context.Context, driver StreamingDeviceDriver, ifaceName string) (StreamHandle, error)
	StreamResource(ctx context.Context, driver StreamingDeviceDriver, interval string) (StreamHandle, error)
	StreamInterfaces(ctx context.Context, driver StreamingDeviceDriver, typeFilter, nameFilter, interval string) (StreamHandle, error)
	StreamIdentity(ctx context.Context, driver StreamingDeviceDriver, interval string) (StreamHandle, error)
	StreamPing(ctx context.Context, driver StreamingDeviceDriver, target string) (StreamHandle, error)
	StreamLogs(ctx context.Context, driver StreamingDeviceDriver, topics string) (StreamHandle, error)
	StreamQueueStats(ctx context.Context, driver StreamingDeviceDriver, params QueueStreamParams) (StreamHandle, error)
	StreamPPPActive(ctx context.Context, driver StreamingDeviceDriver, nameFilter string) (StreamHandle, error)
	StreamPPPSecrets(ctx context.Context, driver StreamingDeviceDriver, nameFilter string) (StreamHandle, error)
	StreamHotspotActive(ctx context.Context, driver StreamingDeviceDriver, userFilter string) (StreamHandle, error)
	StreamHotspotActiveStats(ctx context.Context, driver StreamingDeviceDriver, interval string) (StreamHandle, error)
	StreamClock(ctx context.Context, driver StreamingDeviceDriver, interval string) (StreamHandle, error)
	StreamRouterboard(ctx context.Context, driver StreamingDeviceDriver, interval string) (StreamHandle, error)
	StreamHealth(ctx context.Context, driver StreamingDeviceDriver, interval string) (StreamHandle, error)

	ParseInterfaceTraffic(res command.Result) InterfaceTrafficStats
	ParseResource(res command.Result) SystemResource
	ParseInterfaces(res command.Result) []Interface
	ParseIdentity(res command.Result) SystemIdentity
	ParseLogEntries(res command.Result) []LogEntry
	ParseQueues(res command.Result) []SimpleQueue
	ParsePPPActive(res command.Result) []PPPActiveSession
	ParseHotspotActive(res command.Result) []HotspotActiveSession
	ParseHotspotActiveStats(res command.Result) []HotspotActiveStat
	ParseClock(res command.Result) SystemClock
	ParseRouterboard(res command.Result) SystemRouterboard
	ParseHealth(res command.Result) SystemHealth
}
