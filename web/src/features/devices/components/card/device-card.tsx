import type { DeviceCardProps } from '../../types'
import { useDeviceStatusStream } from '../../hooks/use-device-status-stream'
import { useDeviceTrafficStream } from '../../hooks/use-device-traffic-stream'
import { useDevicePingStream } from '../../hooks/use-device-ping-stream'
import { DeviceCardHeader } from './device-card-header'
import { DeviceCardMetrics } from './device-card-metrics'
import { DeviceCardPing } from './device-card-ping'
import { DeviceCardTraffic } from './device-card-traffic'

export function DeviceCard({ device }: DeviceCardProps) {
  const pingTarget = device.pingTarget || device.extra?.ping_target || '8.8.8.8'

  // 1. Live Status & Interfaces Stream
  const {
    isOnline,
    boardName,
    cpuUsage,
    memUsage,
    uptime,
    version,
    interfaces,
    selectedIface,
    setSelectedIface,
  } = useDeviceStatusStream({ device })

  // 2. Real-Time Traffic Stream
  const { rxBps, txBps, trafficHistory } = useDeviceTrafficStream({
    deviceId: device.id,
    enabled: device.enabled,
    selectedIface,
  })

  // 3. Real-Time Ping Stream
  const { pingMs, pingHistory } = useDevicePingStream({
    deviceId: device.id,
    enabled: device.enabled,
    target: pingTarget,
  })

  return (
    <article className='flex flex-col rounded-xl border border-border/70 bg-card/95 backdrop-blur-xs p-4 text-card-foreground shadow-xs transition-all duration-200 hover:shadow-md hover:border-border'>
      {/* 1. Header & Quick Actions */}
      <DeviceCardHeader
        device={device}
        isOnline={isOnline}
        boardName={boardName}
      />

      {/* 2. Hardware Resource Metrics */}
      <DeviceCardMetrics
        cpuUsage={cpuUsage}
        memUsage={memUsage}
        uptime={uptime}
        version={version}
      />

      {/* 3. Ping Latency & Sparkline */}
      <DeviceCardPing
        device={device}
        pingMs={pingMs}
        pingHistory={pingHistory}
        pingTarget={pingTarget}
      />

      {/* 4. Interface Bandwidth & Real-Time Traffic Chart */}
      <DeviceCardTraffic
        interfaces={interfaces}
        selectedIface={selectedIface}
        setSelectedIface={setSelectedIface}
        rxBps={rxBps}
        txBps={txBps}
        trafficHistory={trafficHistory}
      />
    </article>
  )
}

