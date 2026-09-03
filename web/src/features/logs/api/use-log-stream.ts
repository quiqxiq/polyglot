import { useEffect, useRef, useState, useCallback } from 'react'
import { networkMonitorClient } from '@/lib/api-client'
import { LogItem } from '../types'
import { classifySeverity } from '../lib/log-formatter'

interface UseLogStreamOptions {
  deviceId: string
  topics?: string
  enabled?: boolean
  maxBuffer?: number
}

export function useLogStream({
  deviceId,
  topics = '',
  enabled = true,
  maxBuffer = 1000,
}: UseLogStreamOptions) {
  const [logs, setLogs] = useState<LogItem[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isStreaming, setIsStreaming] = useState(false)
  const [isPaused, setIsPaused] = useState(false)
  const [error, setError] = useState<Error | null>(null)

  const isPausedRef = useRef(isPaused)
  isPausedRef.current = isPaused

  const pausedQueueRef = useRef<LogItem[]>([])
  const seenFingerprintsRef = useRef<Set<string>>(new Set())

  const clearLogs = useCallback(() => {
    console.log('[MikroTik Log Stream] Buffer cleared by user')
    setLogs([])
    pausedQueueRef.current = []
    seenFingerprintsRef.current.clear()
  }, [])

  const togglePause = useCallback(() => {
    setIsPaused((prev) => {
      const next = !prev
      console.log(`[MikroTik Log Stream] Stream ${next ? 'PAUSED' : 'RESUMED'}`)
      if (!next && pausedQueueRef.current.length > 0) {
        // Flush accumulated queue on unpause
        setLogs((current) => {
          const combined = [...current, ...pausedQueueRef.current]
          pausedQueueRef.current = []
          return combined.length > maxBuffer ? combined.slice(combined.length - maxBuffer) : combined
        })
      }
      return next
    })
  }, [maxBuffer])

  useEffect(() => {
    if (!deviceId || !enabled) {
      setLogs([])
      pausedQueueRef.current = []
      seenFingerprintsRef.current.clear()
      setIsLoading(false)
      setIsStreaming(false)
      return
    }

    seenFingerprintsRef.current.clear()
    const abortController = new AbortController()

    async function startStream() {
      console.log('[MikroTik Log Stream] Connecting to stream...', { deviceId, topics })
      while (!abortController.signal.aborted) {
        try {
          setIsLoading(true)
          setError(null)

          const stream = networkMonitorClient.streamLogs(
            { deviceId, topics },
            { signal: abortController.signal }
          )

          setIsStreaming(true)
          setIsLoading(false)

          for await (const frame of stream) {
            if (abortController.signal.aborted) break


            if (frame.logs && frame.logs.length > 0) {
              const newItems: LogItem[] = []

              for (let i = 0; i < frame.logs.length; i++) {
                const item = frame.logs[i]
                const fingerprint = `${item.id || ''}|${item.time || ''}|${item.topics || ''}|${item.message || ''}`

                if (seenFingerprintsRef.current.has(fingerprint)) {
                  continue
                }
                seenFingerprintsRef.current.add(fingerprint)

                if (seenFingerprintsRef.current.size > maxBuffer * 2) {
                  const itemsArray = Array.from(seenFingerprintsRef.current)
                  seenFingerprintsRef.current = new Set(itemsArray.slice(itemsArray.length - maxBuffer))
                }

                const uniqueKey = item.id
                  ? `${item.id}-${Date.now()}-${i}-${Math.random().toString(36).substring(2, 6)}`
                  : `log-${Date.now()}-${i}-${Math.random().toString(36).substring(2, 6)}`

                newItems.push({
                  id: uniqueKey,
                  time: item.time || new Date().toLocaleTimeString(),
                  topics: item.topics || '',
                  message: item.message || '',
                  severity: classifySeverity(item.topics, item.message),
                  timestamp: Number(frame.timestampUnix) * 1000 || Date.now(),
                })
              }

              if (newItems.length > 0) {
                if (isPausedRef.current) {
                  pausedQueueRef.current = [...pausedQueueRef.current, ...newItems].slice(-maxBuffer)
                } else {
                  setLogs((prev) => {
                    const combined = [...prev, ...newItems]
                    return combined.length > maxBuffer ? combined.slice(combined.length - maxBuffer) : combined
                  })
                }
              }
            }
          }

          // If stream gracefully ended from server side, wait before reconnecting
          if (!abortController.signal.aborted) {
            await new Promise((resolve) => setTimeout(resolve, 2000))
          }
        } catch (err: unknown) {
          if (abortController.signal.aborted) {
            break
          }
          const errorObj = err instanceof Error ? err : new Error(String(err))
          setError(errorObj)
          setIsStreaming(false)
          setIsLoading(false)

          await new Promise((resolve) => setTimeout(resolve, 2500))
        }
      }
    }

    startStream()

    return () => {
      abortController.abort()
    }
  }, [deviceId, topics, enabled, maxBuffer])

  return {
    logs,
    isLoading,
    isStreaming,
    isPaused,
    error,
    clearLogs,
    togglePause,
  }
}
