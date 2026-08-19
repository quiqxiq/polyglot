import { useEffect, useRef, useState } from 'react'
import { hotspotClient } from '@/lib/api-client'
import type { HotspotActiveSession, HotspotUser } from '@/gen/v1/hotspot_pb'

export function useStreamActiveSessions(
  deviceId: string,
  enabled = true,
  interval = '1s'
) {
  const [sessions, setSessions] = useState<HotspotActiveSession[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)
  const sessionsMapRef = useRef<Map<string, HotspotActiveSession>>(new Map())

  useEffect(() => {
    if (!deviceId || !enabled) {
      setSessions([])
      sessionsMapRef.current.clear()
      setIsLoading(false)
      return
    }

    const abortController = new AbortController()
    setIsLoading(true)
    setError(null)

    // 1. Session Presence Stream (Lifecycle via follow)
    async function startSessionStream() {
      try {
        const stream = hotspotClient.streamActiveSessions(
          { deviceId },
          { signal: abortController.signal }
        )
        for await (const frame of stream) {
          if (abortController.signal.aborted) break
          const newMap = new Map<string, HotspotActiveSession>()
          for (const s of frame.sessions) {
            const existing = sessionsMapRef.current.get(s.id)
            if (existing) {
              // Preserve dynamic telemetry stats if already updated
              s.uptime = existing.uptime || s.uptime
              s.bytesIn = existing.bytesIn || s.bytesIn
              s.bytesOut = existing.bytesOut || s.bytesOut
            }
            newMap.set(s.id, s)
          }
          sessionsMapRef.current = newMap
          setSessions(Array.from(newMap.values()))
          setIsLoading(false)
        }
      } catch (err: unknown) {
        if (!abortController.signal.aborted) {
          setError(err as Error)
          setIsLoading(false)
        }
      }
    }

    // 2. Realtime Telemetry Stats Stream (Dynamic counters via stats interval=1s)
    async function startStatsStream() {
      try {
        const stream = hotspotClient.streamActiveStats(
          { deviceId, interval },
          { signal: abortController.signal }
        )
        for await (const frame of stream) {
          if (abortController.signal.aborted) break
          if (sessionsMapRef.current.size === 0) continue
          let changed = false
          for (const st of frame.stats) {
            const sess = sessionsMapRef.current.get(st.id)
            if (sess) {
              sess.uptime = st.uptime
              sess.bytesIn = st.bytesIn
              sess.bytesOut = st.bytesOut
              changed = true
            }
          }
          if (changed) {
            setSessions(Array.from(sessionsMapRef.current.values()))
          }
        }
      } catch {
        // Fall back gracefully if stats stream encounters network blip
      }
    }

    startSessionStream()
    startStatsStream()

    return () => {
      abortController.abort()
    }
  }, [deviceId, enabled, interval])

  return { sessions, isLoading, error }
}

export function useStreamHotspotInactive(deviceId: string, enabled = true) {
  const [users, setUsers] = useState<HotspotUser[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  useEffect(() => {
    if (!deviceId || !enabled) {
      setUsers([])
      setIsLoading(false)
      return
    }

    const abortController = new AbortController()
    setIsLoading(true)
    setError(null)

    async function startStream() {
      try {
        const stream = hotspotClient.streamHotspotInactive(
          { deviceId, interval: '1s' },
          { signal: abortController.signal }
        )
        for await (const frame of stream) {
          if (abortController.signal.aborted) break
          setUsers(frame.users)
          setIsLoading(false)
        }
      } catch (err: unknown) {
        if (!abortController.signal.aborted) {
          setError(err as Error)
          setIsLoading(false)
        }
      }
    }

    startStream()

    return () => {
      abortController.abort()
    }
  }, [deviceId, enabled])

  return { users, isLoading, error }
}
