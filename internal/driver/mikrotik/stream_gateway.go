package mikrotik

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/hotspot"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/iface"
	mikrotikppp "github.com/quixiq/polyglot/internal/driver/mikrotik/ppp"
	mikrotikqueue "github.com/quixiq/polyglot/internal/driver/mikrotik/queue"
	mikrotiksystem "github.com/quixiq/polyglot/internal/driver/mikrotik/system"
	"github.com/quixiq/polyglot/internal/port"
)

type monitorStreamGateway struct{}

// NewMonitorStreamGateway creates a port.MonitorStreamGateway implementation backed by MikroTik drivers.
func NewMonitorStreamGateway() port.MonitorStreamGateway {
	return &monitorStreamGateway{}
}

func (g *monitorStreamGateway) StreamTraffic(ctx context.Context, driver port.StreamingDeviceDriver, ifaceName string) (port.StreamHandle, error) {
	return driver.Stream(ctx, iface.NewMonitorTrafficStreamCommand(ifaceName))
}

func (g *monitorStreamGateway) StreamResource(ctx context.Context, driver port.StreamingDeviceDriver, interval string) (port.StreamHandle, error) {
	return driver.Stream(ctx, mikrotiksystem.NewStreamResourceCommand(interval))
}

func (g *monitorStreamGateway) StreamInterfaces(ctx context.Context, driver port.StreamingDeviceDriver, typeFilter, nameFilter, interval string) (port.StreamHandle, error) {
	return driver.Stream(ctx, iface.NewStreamInterfacesCommand(typeFilter, nameFilter, interval))
}

func (g *monitorStreamGateway) StreamIdentity(ctx context.Context, driver port.StreamingDeviceDriver, interval string) (port.StreamHandle, error) {
	return driver.Stream(ctx, mikrotiksystem.NewStreamIdentityCommand(interval))
}

func (g *monitorStreamGateway) StreamPing(ctx context.Context, driver port.StreamingDeviceDriver, target string) (port.StreamHandle, error) {
	return driver.Stream(ctx, mikrotiksystem.NewPingStreamCommand(target))
}

func (g *monitorStreamGateway) StreamLogs(ctx context.Context, driver port.StreamingDeviceDriver, topics string) (port.StreamHandle, error) {
	return driver.Stream(ctx, mikrotiksystem.NewStreamLogsCommand(topics))
}

func (g *monitorStreamGateway) StreamQueueStats(ctx context.Context, driver port.StreamingDeviceDriver, params port.QueueStreamParams) (port.StreamHandle, error) {
	return driver.Stream(ctx, mikrotikqueue.NewStreamQueueStatsCommand(params))
}

func (g *monitorStreamGateway) StreamPPPActive(ctx context.Context, driver port.StreamingDeviceDriver, nameFilter string) (port.StreamHandle, error) {
	return driver.Stream(ctx, mikrotikppp.NewStreamActiveCommand(nameFilter))
}

func (g *monitorStreamGateway) StreamPPPActiveStats(ctx context.Context, driver port.StreamingDeviceDriver, interval string) (port.StreamHandle, error) {
	return driver.Stream(ctx, mikrotikppp.NewStreamActiveStatsCommand(interval))
}

func (g *monitorStreamGateway) StreamHotspotActive(ctx context.Context, driver port.StreamingDeviceDriver, userFilter string) (port.StreamHandle, error) {
	return driver.Stream(ctx, hotspot.NewStreamActiveCommand(userFilter))
}

func (g *monitorStreamGateway) StreamHotspotActiveStats(ctx context.Context, driver port.StreamingDeviceDriver, interval string) (port.StreamHandle, error) {
	return driver.Stream(ctx, hotspot.NewStreamActiveStatsCommand(interval))
}

func (g *monitorStreamGateway) StreamClock(ctx context.Context, driver port.StreamingDeviceDriver, interval string) (port.StreamHandle, error) {
	return driver.Stream(ctx, mikrotiksystem.NewStreamClockCommand(interval))
}

func (g *monitorStreamGateway) StreamRouterboard(ctx context.Context, driver port.StreamingDeviceDriver, interval string) (port.StreamHandle, error) {
	return driver.Stream(ctx, mikrotiksystem.NewStreamRouterboardCommand(interval))
}

func (g *monitorStreamGateway) StreamHealth(ctx context.Context, driver port.StreamingDeviceDriver, interval string) (port.StreamHandle, error) {
	return driver.Stream(ctx, mikrotiksystem.NewStreamHealthCommand(interval))
}

func (g *monitorStreamGateway) ParseInterfaceTraffic(res command.Result) port.InterfaceTrafficStats {
	return iface.ParseInterfaceTrafficStats(res)
}

func (g *monitorStreamGateway) ParseResource(res command.Result) port.SystemResource {
	return mikrotiksystem.ParseResource(res)
}

func (g *monitorStreamGateway) ParseInterfaces(res command.Result) []port.Interface {
	return iface.ParseInterfaces(res)
}

func (g *monitorStreamGateway) ParseIdentity(res command.Result) port.SystemIdentity {
	return mikrotiksystem.ParseIdentity(res)
}

func (g *monitorStreamGateway) ParseLogEntries(res command.Result) []port.LogEntry {
	return mikrotiksystem.ParseLogs(res)
}

func (g *monitorStreamGateway) ParseQueues(res command.Result) []port.SimpleQueue {
	return mikrotikqueue.ParseSimpleQueues(res)
}

func (g *monitorStreamGateway) ParsePPPActive(res command.Result) []port.PPPActiveSession {
	return mikrotikppp.ParseActiveSessions(res)
}

func (g *monitorStreamGateway) ParsePPPActiveStats(res command.Result) []port.PPPActiveStat {
	return mikrotikppp.ParseActiveStats(res)
}

func (g *monitorStreamGateway) ParseHotspotActive(res command.Result) []port.HotspotActiveSession {
	return hotspot.ParseActiveSessions(res)
}

func (g *monitorStreamGateway) ParseHotspotActiveStats(res command.Result) []port.HotspotActiveStat {
	return hotspot.ParseActiveStats(res)
}

func (g *monitorStreamGateway) ParseClock(res command.Result) port.SystemClock {
	return mikrotiksystem.ParseClock(res)
}

func (g *monitorStreamGateway) ParseRouterboard(res command.Result) port.SystemRouterboard {
	return mikrotiksystem.ParseRouterboard(res)
}

func (g *monitorStreamGateway) ParseHealth(res command.Result) port.SystemHealth {
	return mikrotiksystem.ParseHealth(res)
}
