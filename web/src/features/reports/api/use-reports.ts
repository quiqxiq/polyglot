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

// ─── Laporan operasional lanjutan (planning v2 Fase 5) ──────────────────

export function useARAgingQuery(asOf = '') {
  return useQuery({
    queryKey: [...reportsKeys.all, 'ar-aging', asOf] as const,
    queryFn: async () => {
      const res = await reportClient.aRAging({ asOf })
      return res
    },
  })
}

export function useGatewayReconciliationQuery(from = '', to = '') {
  return useQuery({
    queryKey: [...reportsKeys.all, 'gateway-recon', from, to] as const,
    queryFn: async () => {
      const res = await reportClient.gatewayReconciliation({ from, to })
      return res
    },
  })
}

export function usePSBStatsQuery(from = '', to = '') {
  return useQuery({
    queryKey: [...reportsKeys.all, 'psb-stats', from, to] as const,
    queryFn: async () => {
      const res = await reportClient.pSBStats({ from, to })
      return res
    },
  })
}

// ─── Laporan Keuangan Billing Ops v3 ────────────────────────────────────

export function useTaxReportQuery(period = '') {
  return useQuery({
    queryKey: [...reportsKeys.all, 'tax', period] as const,
    queryFn: async () => await reportClient.taxReport({ period }),
  })
}

export function useSalesReportQuery(period = '') {
  return useQuery({
    queryKey: [...reportsKeys.all, 'sales', period] as const,
    queryFn: async () => await reportClient.salesReport({ period }),
  })
}

export function useCollectionReportQuery(period = '') {
  return useQuery({
    queryKey: [...reportsKeys.all, 'collection', period] as const,
    queryFn: async () => await reportClient.collectionReport({ period }),
  })
}

export function useCashflowReportQuery(from = '', to = '') {
  return useQuery({
    queryKey: [...reportsKeys.all, 'cashflow', from, to] as const,
    queryFn: async () => await reportClient.cashflowReport({ from, to }),
  })
}

export function useAgingDetailQuery(asOf = '', limit = 200) {
  return useQuery({
    queryKey: [...reportsKeys.all, 'aging-detail', asOf, limit] as const,
    queryFn: async () => await reportClient.agingDetail({ asOf, limit }),
  })
}

export function useClosingReportQuery(period = '') {
  return useQuery({
    queryKey: [...reportsKeys.all, 'closing', period] as const,
    queryFn: async () => await reportClient.closingReport({ period }),
  })
}

export function useExportReportMutation() {
  return useMutation({
    mutationFn: async (req: {
      reportType: string
      format?: string
      params?: Record<string, string>
    }) =>
      await reportClient.exportReport({
        reportType: req.reportType,
        format: req.format || 'xlsx',
        params: req.params || {},
      }),
  })
}
