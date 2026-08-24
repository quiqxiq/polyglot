import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { hotspotClient, reportClient } from '@/lib/api-client'
import { reportsKeys } from './keys'
import {
  ListHotspotReportsRequest,
  DeleteHotspotReportRequest,
} from '@/gen/v1/hotspot_pb'

// ─── ISP Financial Reports (ReportService) ──────────────────────────────

export function useDailyReportQuery(date = '') {
  return useQuery({
    queryKey: reportsKeys.daily(date),
    queryFn: async () => {
      const res = await reportClient.dailyReport({ date })
      return res.summary
    },
  })
}

export function useMonthlyReportQuery(month = '') {
  return useQuery({
    queryKey: reportsKeys.monthly(month),
    queryFn: async () => {
      const res = await reportClient.monthlyReport({ month })
      return res.summary
    },
  })
}

export function useYearlyReportQuery(year = 0) {
  return useQuery({
    queryKey: reportsKeys.yearly(year),
    queryFn: async () => {
      const res = await reportClient.yearlyReport({ year })
      return res.summary
    },
  })
}

export function useRefreshSnapshotMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (date: string) => {
      return await reportClient.refreshSnapshot({ date })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: reportsKeys.all })
    },
  })
}

// ─── Hotspot Reports (Kompatibilitas HotspotService) ─────────────────────

export function useHotspotReportsQuery(
  deviceId: string,
  day = '',
  month = '',
  year = '',
  enabled = true
) {
  return useQuery({
    queryKey: reportsKeys.hotspot(deviceId, day, month, year),
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
