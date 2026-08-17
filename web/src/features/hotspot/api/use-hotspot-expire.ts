import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { hotspotClient } from '@/lib/api-client'
import { hotspotKeys } from './keys'
import {
  SetupExpireMonitorRequest,
  DisableExpireMonitorRequest,
  RemoveExpireMonitorRequest,
  GetExpireMonitorStatusRequest,
} from '@/gen/v1/hotspot_pb'

export function useExpireMonitorStatusQuery(deviceId: string, enabled = true) {
  return useQuery({
    queryKey: hotspotKeys.expireMonitorStatus(deviceId),
    queryFn: async () => {
      const res = await hotspotClient.getExpireMonitorStatus(
        new GetExpireMonitorStatusRequest({ deviceId })
      )
      return res
    },
    enabled: Boolean(deviceId) && enabled,
  })
}

export function useSetupExpireMonitorMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: { deviceId: string; interval: string }) => {
      return await hotspotClient.setupExpireMonitor(new SetupExpireMonitorRequest(params))
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: hotspotKeys.expireMonitorStatus(variables.deviceId) })
      queryClient.invalidateQueries({ queryKey: hotspotKeys.profiles(variables.deviceId) })
    },
  })
}

export function useDisableExpireMonitorMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: { deviceId: string }) => {
      return await hotspotClient.disableExpireMonitor(new DisableExpireMonitorRequest(params))
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: hotspotKeys.expireMonitorStatus(variables.deviceId) })
    },
  })
}

export function useRemoveExpireMonitorMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: { deviceId: string }) => {
      return await hotspotClient.removeExpireMonitor(new RemoveExpireMonitorRequest(params))
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: hotspotKeys.expireMonitorStatus(variables.deviceId) })
    },
  })
}
