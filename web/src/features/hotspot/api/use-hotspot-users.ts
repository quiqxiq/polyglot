import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { hotspotClient } from '@/lib/api-client'
import { hotspotKeys } from './keys'
import {
  CreateHotspotUserRequest,
  ListHotspotUsersRequest,
  UpdateHotspotUserRequest,
  ResetHotspotUserCountersRequest,
  DeleteHotspotUserRequest,
} from '@/gen/v1/hotspot_pb'

export function useHotspotUsersQuery(
  deviceId: string,
  profile = '',
  comment = '',
  enabled = true
) {
  return useQuery({
    queryKey: hotspotKeys.users(deviceId, profile, comment),
    queryFn: async () => {
      const res = await hotspotClient.listUsers(
        new ListHotspotUsersRequest({ deviceId, profile, comment })
      )
      return res.users
    },
    enabled: Boolean(deviceId) && enabled,
  })
}

export function useHotspotUserQuery(deviceId: string, rosId: string, enabled = true) {
  return useQuery({
    queryKey: hotspotKeys.user(deviceId, rosId),
    queryFn: async () => {
      const res = await hotspotClient.getUser({ deviceId, rosId })
      return res.user
    },
    enabled: Boolean(deviceId) && Boolean(rosId) && enabled,
  })
}

export function useCreateHotspotUserMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: {
      deviceId: string
      name: string
      password: string
      profile: string
      server?: string
      macAddress?: string
      timeLimit?: string
      dataLimit?: string
      comment?: string
    }) => {
      return await hotspotClient.createUser(new CreateHotspotUserRequest(params))
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: hotspotKeys.users(variables.deviceId) })
    },
  })
}

export function useUpdateHotspotUserMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: {
      deviceId: string
      rosId: string
      name: string
      password: string
      profile: string
      server?: string
      macAddress?: string
      timeLimit?: string
      dataLimit?: string
      comment?: string
      resetCounter?: boolean
      expireDate?: string
      userCode?: string
    }) => {
      return await hotspotClient.updateUser(new UpdateHotspotUserRequest(params))
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: hotspotKeys.users(variables.deviceId) })
      if (variables.rosId) {
        queryClient.invalidateQueries({ queryKey: hotspotKeys.user(variables.deviceId, variables.rosId) })
      }
    },
  })
}

export function useResetHotspotUserCountersMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: { deviceId: string; rosId: string }) => {
      return await hotspotClient.resetUserCounters(new ResetHotspotUserCountersRequest(params))
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: hotspotKeys.users(variables.deviceId) })
    },
  })
}

export function useDeleteHotspotUserMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: { deviceId: string; rosId: string }) => {
      return await hotspotClient.deleteUser(new DeleteHotspotUserRequest(params))
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: hotspotKeys.users(variables.deviceId) })
    },
  })
}
