import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { hotspotClient } from '@/lib/api-client'
import { hotspotKeys } from './keys'
import {
  CreateHotspotProfileRequest,
  UpdateHotspotProfileRequest,
  DeleteHotspotProfileRequest,
  HotspotProfileParams,
} from '@/gen/v1/hotspot_pb'

export function useHotspotProfilesQuery(deviceId: string, enabled = true) {
  return useQuery({
    queryKey: hotspotKeys.profiles(deviceId),
    queryFn: async () => {
      const res = await hotspotClient.listProfiles({ deviceId })
      return res.profiles
    },
    enabled: Boolean(deviceId) && enabled,
  })
}

export function useCreateHotspotProfileMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: {
      deviceId: string
      profile: {
        name: string
        addressPool?: string
        sharedUsers?: string
        rateLimit?: string
        parentQueue?: string
        price?: string
        sellingPrice?: string
        validity?: string
        expireMode?: string
        lockUser?: boolean
        lockServer?: boolean
        enableRecording?: boolean
        comment?: string
      }
    }) => {
      return await hotspotClient.createProfile(
        new CreateHotspotProfileRequest({
          deviceId: params.deviceId,
          profile: new HotspotProfileParams(params.profile),
        })
      )
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: hotspotKeys.profiles(variables.deviceId) })
    },
  })
}

export function useUpdateHotspotProfileMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: {
      deviceId: string
      rosId: string
      profile: {
        name: string
        addressPool?: string
        sharedUsers?: string
        rateLimit?: string
        parentQueue?: string
        price?: string
        sellingPrice?: string
        validity?: string
        expireMode?: string
        lockUser?: boolean
        lockServer?: boolean
        enableRecording?: boolean
        comment?: string
      }
    }) => {
      return await hotspotClient.updateProfile(
        new UpdateHotspotProfileRequest({
          deviceId: params.deviceId,
          rosId: params.rosId,
          profile: new HotspotProfileParams(params.profile),
        })
      )
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: hotspotKeys.profiles(variables.deviceId) })
    },
  })
}

export function useDeleteHotspotProfileMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: { deviceId: string; rosId: string }) => {
      return await hotspotClient.deleteProfile(new DeleteHotspotProfileRequest(params))
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: hotspotKeys.profiles(variables.deviceId) })
    },
  })
}
