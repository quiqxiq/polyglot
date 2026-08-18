import { useMutation } from '@tanstack/react-query'
import { hotspotClient } from '@/lib/api-client'

export function useCheckVoucherStatusMutation() {
  return useMutation({
    mutationFn: async (params: { deviceId: string; username: string }) => {
      return await hotspotClient.checkVoucherStatus({
        deviceId: params.deviceId,
        username: params.username,
      })
    },
  })
}
