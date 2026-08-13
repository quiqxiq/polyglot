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
    // Status diperbarui instan via SSE (useWARealtimeStream) — tidak ada
    // polling. staleTime diperpanjang karena SSE menjaga data tetap segar;
    // observer kedua (mis. create-dialog) tidak memicu refetch saat mount.
    staleTime: 15_000,
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
    // QR diperbarui instan via SSE (useWASessionStatusStream) — tidak ada
    // polling. Tombol "Muat Ulang QR" di modal tetap tersedia sebagai
    // fallback manual (mis. koneksi SSE sempat terputus).
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
