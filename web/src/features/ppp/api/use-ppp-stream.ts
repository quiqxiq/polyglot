import { useEffect, useRef, useState } from 'react'
import { pppClient } from '@/lib/api-client'
import type { PPPActiveSession, PPPSecret } from '@/gen/v1/ppp_pb'
import { formatUptime } from '../utils/format-uptime'

export type EnrichedPPPActiveSession = PPPActiveSession & {
  uptime?: string
  limitBytesIn?: string
  limitBytesOut?: string
  rawUptime?: string
  receivedAt?: number
}

export function useStreamPPPActiveSessions(
  deviceId?: string,
  enabled = true,
  _interval = '1s'
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

    // Helper to generate current live sessions array with ticking uptimes
    const getTickedSessions = (): EnrichedPPPActiveSession[] => {
      const now = Date.now()
      const result: EnrichedPPPActiveSession[] = []
      for (const sess of sessionsMapRef.current.values()) {
        const raw = sess.rawUptime || sess.uptime || ''
        const elapsedSec = sess.receivedAt ? Math.floor((now - sess.receivedAt) / 1000) : 0
        const tickedUptime = formatUptime(raw, elapsedSec)

        const cloned: EnrichedPPPActiveSession = sess.clone
          ? sess.clone()
          : Object.assign(Object.create(Object.getPrototypeOf(sess)), sess)
        cloned.uptime = tickedUptime
        cloned.rawUptime = raw
        cloned.receivedAt = sess.receivedAt
        result.push(cloned)
      }
      return result
    }

    // 1. Session Presence Stream (Lifecycle via RouterOS follow)
    async function startSessionStream() {
      try {
        const stream = pppClient.streamActiveSessions(
          { deviceId: deviceId!, nameFilter: '' },
          { signal: abortController.signal }
        )
        for await (const frame of stream) {
          if (abortController.signal.aborted) break
          const newMap = new Map<string, EnrichedPPPActiveSession>()
          const now = Date.now()
          for (const s of frame.sessions) {
            const rawUptime = s.uptime || ''
            const formatted = formatUptime(rawUptime, 0)
            const enriched = Object.assign(s, {
              uptime: formatted,
              rawUptime: rawUptime,
              receivedAt: now,
            }) as EnrichedPPPActiveSession
            newMap.set(s.id, enriched)
          }
          sessionsMapRef.current = newMap
          setSessions(getTickedSessions())
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

    startSessionStream()

    // 2. Local Ticker (1-second interval in browser memory, 0 network requests)
    const tickerInterval = setInterval(() => {
      if (sessionsMapRef.current.size > 0) {
        setSessions(getTickedSessions())
      }
    }, 1000)

    return () => {
      abortController.abort()
      clearInterval(tickerInterval)
    }
  }, [deviceId, enabled])

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

  return { secrets, isLoading, error }
}
