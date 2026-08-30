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
    if (!deviceId || !enabled) {
      return
    }

    const ifaceToStream =
      selectedIface && selectedIface !== 'default' ? selectedIface : 'ether1'

    const controller = new AbortController()
    let active = true
    let retryTimer: ReturnType<typeof setTimeout> | undefined

    const retry = () => {
      if (active && !controller.signal.aborted) {
        retryTimer = setTimeout(() => {
          retryTimer = undefined
          void startTrafficStream()
        }, 3000)
      }
    }

    async function startTrafficStream() {
      try {
        const stream = deviceClient.streamInterfaceTraffic(
          { id: deviceId, interfaceName: ifaceToStream },
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
			if (active) retry()
      } catch (err: unknown) {
        if (!isStreamAbortedError(err, controller.signal, active) && active) {
				retry()
        }
      }
    }

    startTrafficStream()

    return () => {
      active = false
      controller.abort()
	  if (retryTimer) clearTimeout(retryTimer)
    }
  }, [deviceId, enabled, selectedIface, maxHistory])

  return {
    rxBps,
    txBps,
    trafficHistory,
  }
}
