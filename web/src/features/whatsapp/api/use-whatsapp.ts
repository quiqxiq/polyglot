import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { API_BASE_URL } from '@/lib/api-base'
import { useAuthStore } from '@/stores/auth-store'
import { waDeviceKeys } from './keys'
import type { WASession } from '../types'

interface RawSessionDTO {
  id: number
  device_name: string
  phone_number?: string
  jid?: string
  status: string
  connected_at?: string
  created_at?: string
}

function mapSessionDTO(raw: RawSessionDTO): WASession {
  return {
    id: String(raw.id),
    name: raw.device_name,
    phoneNumber: raw.phone_number || '',
    jid: raw.jid || '',
    status: raw.status || 'offline',
    connectedAt: raw.connected_at || '',
    createdAt: raw.created_at || '',
  }
}

async function authFetch(path: string, options: RequestInit = {}) {
  const token = useAuthStore.getState().auth.accessToken
  const headers = new Headers(options.headers)
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }
  if (options.body && typeof options.body === 'string' && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  const res = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers,
  })

  if (!res.ok) {
    const data = await res.json().catch(() => null)
    throw new Error(data?.message || `HTTP ${res.status}: ${res.statusText}`)
  }

  return res.json()
}

export function useWASessionsQuery() {
  return useQuery({
    queryKey: waDeviceKeys.sessions(),
    queryFn: async (): Promise<WASession[]> => {
      const data = (await authFetch('/api/whatsapp/sessions')) as {
        sessions?: RawSessionDTO[]
      }
      return (data.sessions || []).map(mapSessionDTO)
    },
    staleTime: 15_000,
  })
}

export function useCreateWASessionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (vars: {
      name: string
      phoneNumber: string
    }): Promise<{ session: WASession }> => {
      const data = (await authFetch('/api/whatsapp/sessions', {
        method: 'POST',
        body: JSON.stringify({
          device_name: vars.name,
          phone_number: vars.phoneNumber,
        }),
      })) as { session: RawSessionDTO }
      return { session: mapSessionDTO(data.session) }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: waDeviceKeys.sessions() })
    },
  })
}

export function useWASessionQRQuery(sessionId: string, enabled = true) {
  return useQuery({
    queryKey: waDeviceKeys.qr(sessionId),
    queryFn: async (): Promise<{ qrCodeBase64: string }> => {
      const data = (await authFetch(
        `/api/whatsapp/session/qr?session_id=${encodeURIComponent(sessionId)}`
      )) as { qr_code: string }
      return { qrCodeBase64: data.qr_code || '' }
    },
    enabled: Boolean(sessionId) && enabled,
  })
}

export function useGetPairingCodeMutation() {
  return useMutation({
    mutationFn: async (vars: {
      sessionId: string
      phoneNumber: string
    }): Promise<{ pairingCode: string }> => {
      const data = (await authFetch('/api/whatsapp/session/pair-code', {
        method: 'POST',
        body: JSON.stringify({
          session_id: Number(vars.sessionId),
          phone_number: vars.phoneNumber,
        }),
      })) as { pairing_code: string }
      return { pairingCode: data.pairing_code || '' }
    },
  })
}

export function useReconnectWASessionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (vars: { sessionId: string }) => {
      return await authFetch('/api/whatsapp/session/reconnect', {
        method: 'POST',
        body: JSON.stringify({
          session_id: Number(vars.sessionId),
        }),
      })
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: waDeviceKeys.sessions() })
      queryClient.invalidateQueries({
        queryKey: waDeviceKeys.qr(variables.sessionId),
      })
    },
  })
}

export function useLogoutWASessionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (vars: { sessionId: string }) => {
      return await authFetch('/api/whatsapp/session/logout', {
        method: 'POST',
        body: JSON.stringify({
          session_id: Number(vars.sessionId),
        }),
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: waDeviceKeys.sessions() })
    },
  })
}

export function usePurgeWASessionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (vars: { sessionId: string }) => {
      return await authFetch('/api/whatsapp/session/purge', {
        method: 'POST',
        body: JSON.stringify({
          session_id: Number(vars.sessionId),
        }),
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: waDeviceKeys.sessions() })
    },
  })
}
