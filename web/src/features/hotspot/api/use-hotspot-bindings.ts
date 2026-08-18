import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { hotspotClient } from '@/lib/api-client'
import { hotspotKeys } from './keys'

export function useHotspotIPBindingsQuery(deviceId: string, enabled = true) {
  return useQuery({
    queryKey: hotspotKeys.ipBindings(deviceId),
    queryFn: async () => {
      const res = await hotspotClient.listHotspotIPBindings({ deviceId })
      return res.bindings
    },
    enabled: Boolean(deviceId) && enabled,
  })
}

export function useCreateHotspotIPBindingMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: {
      deviceId: string
      macAddress?: string
      address?: string
      toAddress?: string
      server?: string
      type?: string
      comment?: string
      disabled?: boolean
    }) => {
      return await hotspotClient.createHotspotIPBinding({
        deviceId: params.deviceId,
        macAddress: params.macAddress || '',
        address: params.address || '',
        toAddress: params.toAddress || '',
        server: params.server || '',
        type: params.type || 'bypassed',
        comment: params.comment || '',
        disabled: params.disabled ?? false,
      })
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: hotspotKeys.ipBindings(variables.deviceId) })
      queryClient.invalidateQueries({ queryKey: hotspotKeys.hosts(variables.deviceId) })
    },
  })
}

export function useUpdateHotspotIPBindingMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: {
      deviceId: string
      rosId: string
      macAddress?: string
      address?: string
      toAddress?: string
      server?: string
      type?: string
      comment?: string
      disabled?: boolean
    }) => {
      return await hotspotClient.updateHotspotIPBinding({
        deviceId: params.deviceId,
        rosId: params.rosId,
        macAddress: params.macAddress || '',
        address: params.address || '',
        toAddress: params.toAddress || '',
        server: params.server || '',
        type: params.type || 'bypassed',
        comment: params.comment || '',
        disabled: params.disabled ?? false,
      })
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: hotspotKeys.ipBindings(variables.deviceId) })
    },
  })
}

export function useDeleteHotspotIPBindingMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: { deviceId: string; rosId: string }) => {
      return await hotspotClient.deleteHotspotIPBinding({
        deviceId: params.deviceId,
        rosId: params.rosId,
      })
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: hotspotKeys.ipBindings(variables.deviceId) })
    },
  })
}
