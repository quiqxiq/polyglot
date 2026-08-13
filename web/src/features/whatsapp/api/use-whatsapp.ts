import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { whatsappClient } from '@/lib/api-client'
import { waDeviceKeys } from './keys'
import {
  CreateWASessionRequest,
  GetWASessionPairingRequest,
  ReconnectWASessionRequest,
  LogoutWASessionRequest,
  PurgeWASessionRequest,
} from '@/gen/v1/whatsapp_pb'

export function useWASessionsQuery() {
  return useQuery({
    queryKey: waDeviceKeys.sessions(),
    queryFn: async () => {
      const res = await whatsappClient.listSessions({})
      return res.sessions
    },
    // Polling dinamis: cepat (2s) saat ada device belum online (QR/connecting)
    // supaya status terasa live; melambat (10s) saat semua device sudah online.
    refetchInterval: (query) => {
      const sessions = query.state.data ?? []
      const hasPending = sessions.some((s) => s.status !== 'online')
      return hasPending ? 2_000 : 10_000
    },
  })
}

export function useCreateWASessionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (vars: { name: string; phoneNumber: string }) => {
      return await whatsappClient.createSession(
        new CreateWASessionRequest({
          name: vars.name,
          phoneNumber: vars.phoneNumber,
        }),
      )
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: waDeviceKeys.sessions() })
    },
  })
}

export function useWASessionQRQuery(sessionId: string, enabled = true) {
  return useQuery({
    queryKey: waDeviceKeys.qr(sessionId),
    queryFn: async () => {
      const res = await whatsappClient.getQRCode({ sessionId })
      return res
    },
    enabled: Boolean(sessionId) && enabled,
    // QR berubah-ubah setiap beberapa detik selama menunggu scan. Polling
    // tetap berjalan walau QR kosong — backend me-restart aliran QR otomatis
    // saat timeout, sehingga QR baru muncul tanpa klik manual.
    refetchInterval: enabled ? 5_000 : false,
  })
}

export function useGetPairingCodeMutation() {
  return useMutation({
    mutationFn: async (vars: { sessionId: string; phoneNumber: string }) => {
      return await whatsappClient.getPairingCode(
        new GetWASessionPairingRequest({
          sessionId: vars.sessionId,
          phoneNumber: vars.phoneNumber,
        }),
      )
    },
  })
}

export function useReconnectWASessionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (vars: { sessionId: string }) => {
      return await whatsappClient.reconnectSession(
        new ReconnectWASessionRequest({ sessionId: vars.sessionId }),
      )
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: waDeviceKeys.sessions() })
      queryClient.invalidateQueries({ queryKey: waDeviceKeys.qr(variables.sessionId) })
    },
  })
}

export function useLogoutWASessionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (vars: { sessionId: string }) => {
      return await whatsappClient.logoutSession(
        new LogoutWASessionRequest({ sessionId: vars.sessionId }),
      )
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
      return await whatsappClient.purgeSession(
        new PurgeWASessionRequest({ sessionId: vars.sessionId }),
      )
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: waDeviceKeys.sessions() })
    },
  })
}
