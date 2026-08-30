import { useEffect, useState } from 'react'
import type { Device } from '@/gen/v1/device_pb'
import { deviceClient } from '@/lib/api-client'
import type { DeviceInterfaceItem } from '../types'
import { isStreamAbortedError } from '../lib/stream-utils'

interface UseDeviceStatusStreamOptions {
  device: Device
}

export function useDeviceStatusStream({ device }: UseDeviceStatusStreamOptions) {
  const [isOnline, setIsOnline] = useState<boolean>(device.enabled)
  const [boardName, setBoardName] = useState<string>(device.vendor)
  const [cpuUsage, setCpuUsage] = useState<number>(0)
  const [memUsage, setMemUsage] = useState<number>(0)
  const [uptime, setUptime] = useState<string>('N/A')
  const [version, setVersion] = useState<string>('N/A')
  const [interfaces, setInterfaces] = useState<DeviceInterfaceItem[]>([])
  const [selectedIface, setSelectedIface] = useState<string>('default')

  useEffect(() => {
    if (!device.id || !device.enabled) return
    const controller = new AbortController()
    let active = true
    let retryTimer: ReturnType<typeof setTimeout> | undefined

    const retry = () => {
      if (active && !controller.signal.aborted) {
        retryTimer = setTimeout(() => {
          retryTimer = undefined
          void startStatusStream()
        }, 3000)
      }
    }

    async function startStatusStream() {
      try {
        const stream = deviceClient.streamDeviceStatus(
          {
            id: device.id,
            selectedInterface: 'default',
            interfaceTypeFilter: 'ether',
          },
          { signal: controller.signal }
        )

        for await (const frame of stream) {
          if (!active) break
          const res = frame.test
          if (res) {
            setIsOnline(res.status === 'connected' || res.status === 'online')
            if (res.boardName) setBoardName(res.boardName)
            if (res.uptime) setUptime(res.uptime)
            if (res.version) setVersion(res.version)
            if (res.cpuLoad !== undefined) setCpuUsage(res.cpuLoad)

            if (res.totalMemory && res.totalMemory > 0n) {
              const free = Number(res.freeMemory)
              const total = Number(res.totalMemory)
              const usedPct = Math.round(((total - free) / total) * 100)
              setMemUsage(usedPct)
            }

            if (res.interfaceList && res.interfaceList.length > 0) {
              const items: DeviceInterfaceItem[] = res.interfaceList.map((ifc) => ({
                id: ifc.name,
                name: ifc.name,
                type:
                  ifc.type ||
                  (ifc.name.startsWith('ether')
                    ? 'ether'
                    : ifc.name.startsWith('wlan')
                    ? 'wlan'
                    : 'bridge'),
                disabled: Boolean(ifc.disabled),
                running: Boolean(ifc.running),
              }))
              setInterfaces(items)
              setSelectedIface((prev) => {
                if (prev === 'default') {
                  const firstEnabled =
                    items.find((i) => !i.disabled)?.name || items[0]?.name || 'default'
                  return firstEnabled
                }
                return prev
              })
            } else if (res.interfaces && res.interfaces.length > 0) {
              const items: DeviceInterfaceItem[] = res.interfaces.map((name) => ({
                id: name,
                name: name,
                type: name.startsWith('ether')
                  ? 'ether'
                  : name.startsWith('wlan')
                  ? 'wlan'
                  : 'bridge',
                disabled: false,
                running: true,
              }))
              setInterfaces(items)
              setSelectedIface((prev) => (prev === 'default' ? res.interfaces[0] : prev))
            }
          }
        }
			if (active) retry()
      } catch (err: unknown) {
        if (!isStreamAbortedError(err, controller.signal, active) && active) {
          setIsOnline(false)
			retry()
        }
      }
    }

    startStatusStream()

    return () => {
      active = false
      controller.abort()
	  if (retryTimer) clearTimeout(retryTimer)
    }
  }, [device.id, device.enabled])

  return {
    isOnline,
    boardName,
    cpuUsage,
    memUsage,
    uptime,
    version,
    interfaces,
    selectedIface,
    setSelectedIface,
  }
}
