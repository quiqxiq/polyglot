import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { hotspotClient } from '@/lib/api-client'
import { hotspotKeys } from './keys'
import {
  GenerateVouchersRequest,
  RenderVouchersRequest,
  GetVoucherBatchRequest,
} from '@/gen/v1/hotspot_pb'

export function useVoucherBatchQuery(deviceId: string, comment: string, enabled = true) {
  return useQuery({
    queryKey: hotspotKeys.voucherBatch(deviceId, comment),
    queryFn: async () => {
      const res = await hotspotClient.getVoucherBatch(
        new GetVoucherBatchRequest({ deviceId, comment })
      )
      return res
    },
    enabled: Boolean(deviceId) && Boolean(comment) && enabled,
  })
}

export function useHotspotTemplatesQuery() {
  return useQuery({
    queryKey: hotspotKeys.templates(),
    queryFn: async () => {
      const res = await hotspotClient.listTemplates({})
      return res.templates
    },
  })
}

export function useGenerateVouchersMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: {
      deviceId: string
      profile: string
      count: number
      userType?: string
      userLength?: number
      prefix?: string
      characterSet?: string
      server?: string
      timeLimit?: string
      dataLimit?: string
      comment?: string
    }) => {
      return await hotspotClient.generateVouchers(new GenerateVouchersRequest(params))
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: hotspotKeys.users(variables.deviceId) })
    },
  })
}

export function useRenderVouchersMutation() {
  return useMutation({
    mutationFn: async (params: {
      deviceId: string
      templateName: string
      comment?: string
      userId?: string
      preview?: boolean
    }) => {
      return await hotspotClient.renderVouchers(new RenderVouchersRequest(params))
    },
  })
}
