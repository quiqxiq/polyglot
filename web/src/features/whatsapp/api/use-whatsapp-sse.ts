import { useEffect, useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { API_BASE_URL } from '@/lib/api-base'
import type { WARealtimeStatus } from '@/lib/realtime'
import { waDeviceKeys } from './keys'
import type { SessionStatusPayload, WASession } from '../types'

const SSE_BASE = API_BASE_URL

export type WARealtimeState = {
  status: WARealtimeStatus
}

export function useWARealtimeStream(): WARealtimeState {
  const queryClient = useQueryClient()
  const [sseStatus, setSseStatus] = useState<WARealtimeStatus>('connecting')

  useEffect(() => {
    const es = new EventSource(`${SSE_BASE}/events`, { withCredentials: true })
    let hasConnectedOnce = false

    const applyStatus = (raw: string) => {
      let payload: SessionStatusPayload
      try {
        payload = JSON.parse(raw) as SessionStatusPayload
      } catch {
        return
      }

      const sessionId = String(payload.session_id)

      if (payload.qr_code) {
        queryClient.setQueryData(waDeviceKeys.qr(sessionId), {
          qrCodeBase64: payload.qr_code,
        })
      }

      queryClient.setQueryData<WASession[]>(waDeviceKeys.sessions(), (old) => {
        if (!old) return old
        return old.map((s) => {
          if (s.id !== sessionId) return s
          return {
            ...s,
            status: payload.status || s.status,
            jid: payload.jid || s.jid,
            phoneNumber: payload.phone_number || s.phoneNumber,
          }
        })
      })
    }

    const handleSessionStatus = (e: MessageEvent) => applyStatus(e.data)

    es.addEventListener('session_status', handleSessionStatus)
    es.onopen = () => {
      setSseStatus('open')
      if (hasConnectedOnce) {
        queryClient.invalidateQueries({ queryKey: waDeviceKeys.all })
      } else {
        const state = queryClient.getQueryState(waDeviceKeys.sessions())
        if (state?.status === 'error' || state?.dataUpdatedAt === 0) {
          queryClient.invalidateQueries({ queryKey: waDeviceKeys.all })
        }
      }
      hasConnectedOnce = true
    }
    es.onerror = () => {
      setSseStatus(
        es.readyState === EventSource.CLOSED
          ? 'closed'
          : hasConnectedOnce
            ? 'reconnecting'
            : 'connecting'
      )
    }

    return () => {
      es.removeEventListener('session_status', handleSessionStatus)
      es.close()
    }
  }, [queryClient])

  return { status: sseStatus }
}
