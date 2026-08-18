import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { hotspotClient } from '@/lib/api-client'
import { hotspotKeys } from './keys'

export function useHotspotCookiesQuery(deviceId: string, enabled = true) {
  return useQuery({
    queryKey: hotspotKeys.cookies(deviceId),
    queryFn: async () => {
      const res = await hotspotClient.listHotspotCookies({ deviceId })
      return res.cookies
    },
    enabled: Boolean(deviceId) && enabled,
  })
}

export function useDeleteHotspotCookieMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: { deviceId: string; rosId?: string }) => {
      return await hotspotClient.deleteHotspotCookie({
        deviceId: params.deviceId,
        rosId: params.rosId || 'all',
      })
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: hotspotKeys.cookies(variables.deviceId) })
    },
  })
}
