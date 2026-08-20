import { pppClient } from '@/lib/api-client'
import {
  CreatePPPProfileRequest,
  DeletePPPProfileRequest,
  ListPPPProfilesRequest,
  UpdatePPPProfileRequest,
} from '@/gen/v1/ppp_pb'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { pppKeys } from './keys'

export function usePPPProfilesQuery(deviceId?: string, nameFilter?: string) {
  return useQuery({
    queryKey: [...pppKeys.profiles(deviceId), nameFilter],
    queryFn: async () => {
      if (!deviceId) return []
      const res = await pppClient.listProfiles(
        new ListPPPProfilesRequest({
          deviceId,
          nameFilter: nameFilter || '',
        })
      )
      return res.profiles
    },
    enabled: !!deviceId,
  })
}

export type CreatePPPProfileParams = {
  deviceId: string
  name: string
  rateLimit?: string
  localAddress?: string
  remoteAddress?: string
  dnsServer?: string
  parentQueue?: string
  addressList?: string
  comment?: string
  sharedUsers?: string
  onlyOne?: string
  useMpls?: string
  useCompression?: string
  useEncryption?: string
  changeTcpMss?: string
  bridgeLearning?: string
}

export function useCreatePPPProfileMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: CreatePPPProfileParams) => {
      return await pppClient.createProfile(new CreatePPPProfileRequest(params))
    },
    onSuccess: (_, variables) => {
      toast.success('PPP Profile created successfully')
      queryClient.invalidateQueries({
        queryKey: pppKeys.profiles(variables.deviceId),
      })
    },
    onError: (err: Error) => {
      toast.error(`Failed to create profile: ${err.message}`)
    },
  })
}

export type UpdatePPPProfileParams = {
  deviceId: string
  rosId: string
  name?: string
  rateLimit?: string
  localAddress?: string
  remoteAddress?: string
  dnsServer?: string
  parentQueue?: string
  addressList?: string
  comment?: string
  sharedUsers?: string
  onlyOne?: string
  useMpls?: string
  useCompression?: string
  useEncryption?: string
  changeTcpMss?: string
  bridgeLearning?: string
}

export function useUpdatePPPProfileMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: UpdatePPPProfileParams) => {
      return await pppClient.updateProfile(new UpdatePPPProfileRequest(params))
    },
    onSuccess: (_, variables) => {
      toast.success('PPP Profile updated successfully')
      queryClient.invalidateQueries({
        queryKey: pppKeys.profiles(variables.deviceId),
      })
    },
    onError: (err: Error) => {
      toast.error(`Failed to update profile: ${err.message}`)
    },
  })
}

export type DeletePPPProfileParams = {
  deviceId: string
  rosId: string
}

export function useDeletePPPProfileMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: DeletePPPProfileParams) => {
      return await pppClient.deleteProfile(new DeletePPPProfileRequest(params))
    },
    onSuccess: (_, variables) => {
      toast.success('PPP Profile deleted successfully')
      queryClient.invalidateQueries({
        queryKey: pppKeys.profiles(variables.deviceId),
      })
    },
    onError: (err: Error) => {
      toast.error(`Failed to delete profile: ${err.message}`)
    },
  })
}
