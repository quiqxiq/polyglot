import { useEffect, useRef, useState } from 'react'
import { hotspotClient } from '@/lib/api-client'
import type { HotspotActiveSession, HotspotActiveStat, HotspotUser } from '@/gen/v1/hotspot_pb'

export type EnrichedHotspotActiveSession = HotspotActiveSession & {
  uptime?: string
  sessionTimeLeft?: string
  idleTime?: string
  bytesIn?: string
  bytesOut?: string
  packetsIn?: string
  packetsOut?: string
}

export function useStreamActiveSessions(
  deviceId: string,
  enabled = true,
  interval = '1s'
) {
  const [sessions, setSessions] = useState<EnrichedHotspotActiveSession[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)
  const sessionsMapRef = useRef<Map<string, EnrichedHotspotActiveSession>>(new Map())
  const statsMapRef = useRef<Map<string, HotspotActiveStat>>(new Map())

  useEffect(() => {
    if (!deviceId || !enabled) {
      setSessions([])
      sessionsMapRef.current.clear()
      statsMapRef.current.clear()
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
          const newMap = new Map<string, EnrichedHotspotActiveSession>()
          for (const s of frame.sessions) {
            const existing = sessionsMapRef.current.get(s.id)
            const stat = statsMapRef.current.get(s.id)
            const enriched = Object.assign(s, {
              uptime: stat?.uptime || existing?.uptime || '',
              sessionTimeLeft: stat?.sessionTimeLeft || existing?.sessionTimeLeft || '',
              idleTime: stat?.idleTime || existing?.idleTime || '',
              bytesIn: stat?.bytesIn || existing?.bytesIn || '',
              bytesOut: stat?.bytesOut || existing?.bytesOut || '',
              packetsIn: stat?.packetsIn || existing?.packetsIn || '',
              packetsOut: stat?.packetsOut || existing?.packetsOut || '',
            }) as EnrichedHotspotActiveSession
            newMap.set(s.id, enriched)
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
      } finally {
        if (!abortController.signal.aborted) {
          setTimeout(() => {
            if (!abortController.signal.aborted) {
              startSessionStream()
            }
          }, 3000)
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
          let changed = false
          for (const st of frame.stats) {
            statsMapRef.current.set(st.id, st)
            const sess = sessionsMapRef.current.get(st.id)
            if (sess) {
              const updated: EnrichedHotspotActiveSession = Object.assign(sess.clone ? sess.clone() : { ...sess }, {
                uptime: st.uptime || sess.uptime || '',
                sessionTimeLeft: st.sessionTimeLeft || sess.sessionTimeLeft || '',
                idleTime: st.idleTime || sess.idleTime || '',
                bytesIn: st.bytesIn || sess.bytesIn || '',
                bytesOut: st.bytesOut || sess.bytesOut || '',
                packetsIn: st.packetsIn || sess.packetsIn || '',
                packetsOut: st.packetsOut || sess.packetsOut || '',
              }) as EnrichedHotspotActiveSession
              sessionsMapRef.current.set(st.id, updated)
              changed = true
            }
          }
          if (changed) {
            setSessions(Array.from(sessionsMapRef.current.values()))
          }
        }
      } catch {
        // Fall back gracefully if stats stream encounters network blip
      } finally {
        if (!abortController.signal.aborted) {
          setTimeout(() => {
            if (!abortController.signal.aborted) {
              startStatsStream()
            }
          }, 3000)
        }
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
      } finally {
        if (!abortController.signal.aborted) {
          setTimeout(() => {
            if (!abortController.signal.aborted) {
              startStream()
            }
          }, 3000)
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
