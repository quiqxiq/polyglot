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
  const [pingMs, setPingMs] = useState<number>(0)
  const [pingHistory, setPingHistory] = useState<PingDataPoint[]>([])

  useEffect(() => {
    if (!deviceId || !enabled || !target) return
    const controller = new AbortController()
    let active = true

    async function startPingStream() {
      try {
        const stream = deviceClient.streamPing(
          { id: deviceId, address: target },
          { signal: controller.signal }
        )

        for await (const frame of stream) {
          if (!active) break
          const latency = Number(frame.latencyMs)
          if (latency > 0) {
            setPingMs(latency)
            setPingHistory((prev) => {
              const next = [
                ...prev,
                {
                  ms: latency,
                  alive: frame.status !== 'timeout',
                  timestamp: Date.now(),
                },
              ]
              return next.slice(-maxHistory)
            })
          }
        }
      } catch (err: unknown) {
        if (!isStreamAbortedError(err, controller.signal, active) && active) {
          setTimeout(() => {
            if (active && !controller.signal.aborted) {
              startPingStream()
            }
          }, 3000)
        }
      }
    }

    startPingStream()

    return () => {
      active = false
      controller.abort()
    }
  }, [deviceId, enabled, target, maxHistory])

  return {
    pingMs,
    pingHistory,
  }
}

