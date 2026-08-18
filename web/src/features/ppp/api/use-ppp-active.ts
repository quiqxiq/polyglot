import { pppClient } from '@/lib/api-client'
import {
  KickPPPActiveSessionRequest,
  KickPPPActiveSessionsRequest,
  ListPPPActiveSessionsRequest,
} from '@/gen/v1/ppp_pb'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { pppKeys } from './keys'

export function usePPPActiveSessionsQuery(deviceId?: string, nameFilter?: string) {
  return useQuery({
    queryKey: [...pppKeys.active(deviceId), nameFilter],
    queryFn: async () => {
      if (!deviceId) return []
      const res = await pppClient.listActiveSessions(
        new ListPPPActiveSessionsRequest({
          deviceId,
          nameFilter: nameFilter || '',
        })
      )
      return res.sessions
    },
    enabled: !!deviceId,
  })
}

export type KickPPPActiveSessionParams = {
  deviceId: string
  rosId: string
}

export function useKickPPPActiveSessionMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: KickPPPActiveSessionParams) => {
      return await pppClient.kickActiveSession(
        new KickPPPActiveSessionRequest(params)
      )
    },
    onSuccess: (_, variables) => {
      toast.success('Active PPPoE session disconnected')
      queryClient.invalidateQueries({
        queryKey: pppKeys.active(variables.deviceId),
      })
      queryClient.invalidateQueries({
        queryKey: pppKeys.inactive(variables.deviceId),
      })
    },
    onError: (err: Error) => {
      toast.error(`Failed to disconnect session: ${err.message}`)
    },
  })
}

export function useBulkKickPPPActiveSessionsMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({
      deviceId,
      rosIds,
    }: {
      deviceId: string
      rosIds: string[]
    }) => {
      const res = await pppClient.kickActiveSessions(
        new KickPPPActiveSessionsRequest({ deviceId, rosIds })
      )
      return res.kickedCount
    },
    onSuccess: (count, variables) => {
      toast.success(`${count} active sessions disconnected`)
      queryClient.invalidateQueries({
        queryKey: pppKeys.active(variables.deviceId),
      })
      queryClient.invalidateQueries({
        queryKey: pppKeys.inactive(variables.deviceId),
      })
    },
    onError: (err: Error) => {
      toast.error(`Failed to disconnect sessions: ${err.message}`)
    },
  })
}
