import { useEffect, useState } from 'react'
import { pppClient } from '@/lib/api-client'
import type { PPPActiveSession, PPPSecret } from '@/gen/v1/ppp_pb'

export function useStreamPPPActiveSessions(deviceId?: string, enabled = true) {
  const [sessions, setSessions] = useState<PPPActiveSession[]>([])
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
        const stream = pppClient.streamActiveSessions(
          { deviceId: deviceId!, nameFilter: '' },
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
