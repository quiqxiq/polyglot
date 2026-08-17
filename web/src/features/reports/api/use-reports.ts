import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { hotspotClient } from '@/lib/api-client'
import { reportsKeys } from './keys'
import {
  ListHotspotReportsRequest,
  DeleteHotspotReportRequest,
} from '@/gen/v1/hotspot_pb'

export function useHotspotReportsQuery(
  deviceId: string,
  day = '',
  month = '',
  year = '',
  enabled = true
) {
  return useQuery({
    queryKey: reportsKeys.list(deviceId, day, month, year),
    queryFn: async () => {
      const res = await hotspotClient.listReports(
        new ListHotspotReportsRequest({
          deviceId,
          day,
          month,
          year,
          summaryOnly: false,
        })
      )
      return res
    },
    enabled: Boolean(deviceId) && enabled,
  })
}

export function useDeleteHotspotReportMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: { deviceId: string; rosId: string }) => {
      return await hotspotClient.deleteReport(new DeleteHotspotReportRequest(params))
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: reportsKeys.all })
    },
  })
}
