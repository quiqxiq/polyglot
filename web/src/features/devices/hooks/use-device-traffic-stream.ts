import { useEffect, useState } from 'react'
import { deviceClient } from '@/lib/api-client'
import type { TrafficDataPoint } from '../types'
import { isStreamAbortedError } from '../lib/stream-utils'

interface UseDeviceTrafficStreamOptions {
  deviceId: string
  enabled: boolean
  selectedIface: string
  maxHistory?: number
}

const DEFAULT_TRAFFIC_HISTORY_MAX = 60

export function useDeviceTrafficStream({
  deviceId,
  enabled,
  selectedIface,
  maxHistory = DEFAULT_TRAFFIC_HISTORY_MAX,
}: UseDeviceTrafficStreamOptions) {
  const [rxBps, setRxBps] = useState<number>(0)
  const [txBps, setTxBps] = useState<number>(0)
  const [trafficHistory, setTrafficHistory] = useState<TrafficDataPoint[]>([])

  useEffect(() => {
    if (!deviceId || !enabled || !selectedIface || selectedIface === 'default') {
      return
    }

    const controller = new AbortController()
    let active = true

    async function startTrafficStream() {
      // Reset traffic state for the new interface
      setRxBps(0)
      setTxBps(0)
      setTrafficHistory([])

      try {
        const stream = deviceClient.streamInterfaceTraffic(
          { id: deviceId, interfaceName: selectedIface },
          { signal: controller.signal }
        )

        for await (const frame of stream) {
          if (!active) break
          const rx = Number(frame.rxBps || 0)
          const tx = Number(frame.txBps || 0)
          setRxBps(rx)
          setTxBps(tx)
          setTrafficHistory((prev) => {
            const next = [...prev, { time: Date.now(), rx, tx }]
            return next.slice(-maxHistory)
          })
        }
      } catch (err: unknown) {
        if (!isStreamAbortedError(err, controller.signal, active) && active) {
          setTimeout(() => {
            if (active && !controller.signal.aborted) {
              startTrafficStream()
            }
          }, 3000)
        }
      }
    }

    startTrafficStream()

    return () => {
      active = false
      controller.abort()
    }
  }, [deviceId, enabled, selectedIface, maxHistory])

  return {
    rxBps,
    txBps,
    trafficHistory,
  }
}

