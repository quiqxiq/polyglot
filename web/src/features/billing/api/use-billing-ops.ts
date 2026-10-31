import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import {
  type CreateInvoiceItem,
  type PaymentMethod,
  ExportInvoicesRequest,
  ListInvoicesRequest,
} from '@/gen/v1/billing_pb'
import { billingClient } from '@/lib/api-client'
import { billingKeys } from './keys'

export interface InvoiceListFilter {
  customerId?: string
  status?: string
  period?: string
  overdueOnly?: boolean
  unpaidOnly?: boolean
  unpaidDaysDisplay?: number
  search?: string
  limit?: number
  offset?: number
}

export function useInvoiceListQuery(filter: InvoiceListFilter = {}) {
  return useQuery({
    queryKey: [
      ...billingKeys.invoices.all(),
      'filtered',
      filter.customerId || '',
      filter.status || '',
      filter.period || '',
      filter.overdueOnly || false,
      filter.unpaidOnly || false,
      filter.unpaidDaysDisplay || 0,
      filter.search || '',
      filter.limit || 0,
      filter.offset || 0,
    ],
    queryFn: async () => {
      const req = new ListInvoicesRequest({
        customerId: filter.customerId || '',
        status: filter.status || '',
        period: filter.period || '',
        overdueOnly: filter.overdueOnly || false,
        unpaidOnly: filter.unpaidOnly || false,
        unpaidDaysDisplay: filter.unpaidDaysDisplay || 0,
        search: filter.search || '',
        limit: filter.limit || 0,
        offset: filter.offset || 0,
      })
      const res = await billingClient.listInvoices(req)
      return {
        invoices: res.invoices,
        summary: res.summary,
        total: res.total,
      }
    },
  })
}

export function useCreateInvoiceMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (req: {
      customerId: string
      subscriptionId?: string
      period?: string
      dueDateUnix?: bigint
      taxPercent?: number
      items: CreateInvoiceItem[]
      notes?: string
    }) =>
      await billingClient.createInvoice({
        customerId: req.customerId,
        subscriptionId: req.subscriptionId || '',
        period: req.period || '',
        dueDateUnix: req.dueDateUnix || 0n,
        taxPercent: req.taxPercent || 0,
        items: req.items,
        notes: req.notes || '',
      }),
    onSuccess: () => {
      toast.success('Tagihan manual berhasil dibuat')
      queryClient.invalidateQueries({ queryKey: billingKeys.invoices.all() })
    },
    onError: (err: Error) => toast.error(`Gagal membuat tagihan: ${err.message}`),
  })
}

export function useExportInvoicesMutation() {
  return useMutation({
    mutationFn: async (req: {
      filter?: InvoiceListFilter
      format?: string
    }) => {
      const f = req.filter || {}
      const exportReq = new ExportInvoicesRequest({
        format: req.format || 'xlsx',
        filter: new ListInvoicesRequest({
          customerId: f.customerId || '',
          status: f.status || '',
          period: f.period || '',
          overdueOnly: f.overdueOnly || false,
          unpaidOnly: f.unpaidOnly || false,
          search: f.search || '',
        }),
      })
      return await billingClient.exportInvoices(exportReq)
    },
    onSuccess: (res) => {
      downloadBlob(res.payload, res.contentType, res.filename)
    },
    onError: (err: Error) => toast.error(`Gagal ekspor: ${err.message}`),
  })
}

// ─── Payment methods ────────────────────────────────────────────────────────

export function usePaymentMethodsQuery(activeOnly = false) {
  return useQuery({
    queryKey: ['billing', 'payment-methods', activeOnly],
    queryFn: async () => {
      const res = await billingClient.listPaymentMethods({ activeOnly })
      return res.paymentMethods
    },
  })
}

export function useSavePaymentMethodMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (pm: PaymentMethod) =>
      await billingClient.savePaymentMethod({ paymentMethod: pm }),
    onSuccess: () => {
      toast.success('Metode pembayaran disimpan')
      queryClient.invalidateQueries({ queryKey: ['billing', 'payment-methods'] })
    },
    onError: (err: Error) => toast.error(`Gagal menyimpan: ${err.message}`),
  })
}

export function useDeletePaymentMethodMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => await billingClient.deletePaymentMethod({ id }),
    onSuccess: () => {
      toast.success('Metode pembayaran dinonaktifkan')
      queryClient.invalidateQueries({ queryKey: ['billing', 'payment-methods'] })
    },
    onError: (err: Error) => toast.error(`Gagal menonaktifkan: ${err.message}`),
  })
}

// ─── Mass payment ───────────────────────────────────────────────────────────

export function useCreateMassPaymentMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (req: {
      invoiceIds: string[]
      paymentMethodId?: string
      notes?: string
    }) =>
      await billingClient.createMassPayment({
        invoiceIds: req.invoiceIds,
        paymentMethodId: req.paymentMethodId || '',
        notes: req.notes || '',
      }),
    onSuccess: () => {
      toast.success('Batch pembayaran massal dibuat')
      queryClient.invalidateQueries({ queryKey: billingKeys.invoices.all() })
      queryClient.invalidateQueries({ queryKey: ['billing', 'mass-payments'] })
    },
    onError: (err: Error) => toast.error(`Gagal membuat batch: ${err.message}`),
  })
}

export function useMassPaymentsQuery(limit = 20) {
  return useQuery({
    queryKey: ['billing', 'mass-payments', limit],
    queryFn: async () => {
      const res = await billingClient.listMassPayments({ limit })
      return res.batches
    },
  })
}

// ─── Gateway transactions & settlements ────────────────────────────────────

export function useGatewayTransactionsQuery(req: {
  gateway?: string
  status?: string
  invoiceId?: string
  limit?: number
  offset?: number
}) {
  return useQuery({
    queryKey: ['billing', 'gateway-transactions', req],
    queryFn: async () => {
      const res = await billingClient.listGatewayTransactions({
        gateway: req.gateway || '',
        status: req.status || '',
        invoiceId: req.invoiceId || '',
        limit: req.limit || 100,
        offset: req.offset || 0,
      })
      return { transactions: res.transactions, total: res.total }
    },
  })
}

export function useSyncGatewayTransactionsMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async () => await billingClient.syncGatewayTransactions({}),
    onSuccess: (res) => {
      toast.success(`Sinkron selesai: ${res.synced} diproses, ${res.expired} kedaluwarsa`)
      queryClient.invalidateQueries({ queryKey: ['billing', 'gateway-transactions'] })
      queryClient.invalidateQueries({ queryKey: billingKeys.invoices.all() })
    },
    onError: (err: Error) => toast.error(`Sinkron gagal: ${err.message}`),
  })
}

export function useGatewaySettlementsQuery(req: { gateway?: string; status?: string }) {
  return useQuery({
    queryKey: ['billing', 'gateway-settlements', req],
    queryFn: async () => {
      const res = await billingClient.listGatewaySettlements({
        gateway: req.gateway || '',
        status: req.status || '',
      })
      return res.settlements
    },
  })
}

export function useSaveGatewaySettlementMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (settlement: Parameters<typeof billingClient.saveGatewaySettlement>[0]['settlement']) =>
      await billingClient.saveGatewaySettlement({ settlement }),
    onSuccess: () => {
      toast.success('Settlement disimpan')
      queryClient.invalidateQueries({ queryKey: ['billing', 'gateway-settlements'] })
    },
    onError: (err: Error) => toast.error(`Gagal menyimpan settlement: ${err.message}`),
  })
}

// ─── Utilities ──────────────────────────────────────────────────────────────

export function downloadBlob(payload: Uint8Array, contentType: string, filename: string) {
  const blob = new Blob([payload as BlobPart], { type: contentType })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
}
