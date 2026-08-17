import { useEffect, useState } from 'react'
import { hotspotClient } from '@/lib/api-client'
import type { HotspotActiveSession, HotspotUser } from '@/gen/v1/hotspot_pb'

export function useStreamActiveSessions(deviceId: string, enabled = true) {
  const [sessions, setSessions] = useState<HotspotActiveSession[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  useEffect(() => {
    if (!deviceId || !enabled) {
      setSessions([])
      setIsLoading(false)
      return
    }

    const abortController = new AbortController()
    setIsLoading(true)
    setError(null)

    async function startStream() {
      try {
        const stream = hotspotClient.streamActiveSessions(
          { deviceId },
          { signal: abortController.signal }
        )
        for await (const frame of stream) {
          if (abortController.signal.aborted) break
          setSessions(frame.sessions)
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
