import { useEffect, useRef, useState } from 'react'
import { pppClient } from '@/lib/api-client'
import type { PPPActiveSession, PPPSecret } from '@/gen/v1/ppp_pb'

export type EnrichedPPPActiveSession = PPPActiveSession & {
  uptime?: string
  limitBytesIn?: string
  limitBytesOut?: string
  bytesIn?: string
  bytesOut?: string
  packetsIn?: string
  packetsOut?: string
}

export function useStreamPPPActiveSessions(
  deviceId?: string,
  enabled = true,
  interval = '1s'
) {
  const [sessions, setSessions] = useState<EnrichedPPPActiveSession[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)
  const sessionsMapRef = useRef<Map<string, EnrichedPPPActiveSession>>(new Map())

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
        const stream = pppClient.streamActiveSessions(
          { deviceId: deviceId!, nameFilter: '' },
          { signal: abortController.signal }
        )
        for await (const frame of stream) {
          if (abortController.signal.aborted) break
          const newMap = new Map<string, EnrichedPPPActiveSession>()
          for (const s of frame.sessions) {
            const existing = sessionsMapRef.current.get(s.id)
            const enriched = Object.assign(s, {
              uptime: existing?.uptime || '',
              limitBytesIn: existing?.limitBytesIn || '',
              limitBytesOut: existing?.limitBytesOut || '',
              bytesIn: existing?.bytesIn || '',
              bytesOut: existing?.bytesOut || '',
              packetsIn: existing?.packetsIn || '',
              packetsOut: existing?.packetsOut || '',
            }) as EnrichedPPPActiveSession
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
      }
    }

    // 2. Realtime Telemetry Stats Stream (Dynamic counters via stats interval=1s)
    async function startStatsStream() {
      try {
        const stream = pppClient.streamActiveStats(
          { deviceId: deviceId!, interval },
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
              sess.limitBytesIn = st.limitBytesIn
              sess.limitBytesOut = st.limitBytesOut
              sess.bytesIn = st.bytesIn
              sess.bytesOut = st.bytesOut
              sess.packetsIn = st.packetsIn
              sess.packetsOut = st.packetsOut
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

export function useStreamPPPInactiveSecrets(deviceId?: string, enabled = true) {
  const [secrets, setSecrets] = useState<PPPSecret[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  useEffect(() => {
    if (!deviceId || !enabled) {
      setSecrets([])
      setIsLoading(false)
      return
    }

    const abortController = new AbortController()
    setIsLoading(true)
    setError(null)

    async function startStream() {
      try {
        const stream = pppClient.streamInactiveSecrets(
          { deviceId: deviceId! },
          { signal: abortController.signal }
        )
        for await (const frame of stream) {
          if (abortController.signal.aborted) break
          setSecrets(frame.secrets)
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

  return { secrets, isLoading, error }
}
