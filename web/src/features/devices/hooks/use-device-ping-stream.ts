import { useEffect, useState } from 'react'
import { deviceClient } from '@/lib/api-client'
import type { PingDataPoint } from '../types'
import { isStreamAbortedError } from '../lib/stream-utils'

interface UseDevicePingStreamOptions {
  deviceId: string
  enabled: boolean
  target: string
  maxHistory?: number
}

const DEFAULT_PING_HISTORY_MAX = 40

export function useDevicePingStream({
  deviceId,
  enabled,
  target,
  maxHistory = DEFAULT_PING_HISTORY_MAX,
}: UseDevicePingStreamOptions) {
  const [pingMs, setPingMs] = useState<number | null>(null)
  const [pingHistory, setPingHistory] = useState<PingDataPoint[]>([])

  useEffect(() => {
    if (!deviceId || !enabled || !target) return
    const controller = new AbortController()
    let active = true
    let retryTimer: ReturnType<typeof setTimeout> | undefined

    const retry = () => {
      if (active && !controller.signal.aborted) {
        retryTimer = setTimeout(() => {
          retryTimer = undefined
          void startPingStream()
        }, 3000)
      }
    }

    async function startPingStream() {
      try {
        const stream = deviceClient.streamPing(
          { id: deviceId, address: target },
          { signal: controller.signal }
        )

        for await (const frame of stream) {
          if (!active) break
          const latency = Number(frame.latencyMs ?? 0)
          const isAlive =
            frame.status !== 'timeout' &&
            frame.status !== 'host unreachable' &&
            frame.status !== 'net unreachable'

          setPingMs(isAlive ? latency : null)
          setPingHistory((prev) => {
            const next = [
              ...prev,
              {
                ms: latency,
                alive: isAlive,
                timestamp: Date.now(),
              },
            ]
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

    startPingStream()

    return () => {
      active = false
      controller.abort()
	  if (retryTimer) clearTimeout(retryTimer)
    }
  }, [deviceId, enabled, target, maxHistory])

  return {
    pingMs,
    pingHistory,
  }
}
