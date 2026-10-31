import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { debtClient } from '@/lib/api-client'

export const debtKeys = {
  all: ['debts'] as const,
  list: (filter: Record<string, unknown>) => [...debtKeys.all, 'list', filter] as const,
  summary: (customerId?: string) => [...debtKeys.all, 'summary', customerId || 'all'] as const,
}

export interface DebtFilterInput {
  customerId?: string
  status?: string
  search?: string
  overdueOnly?: boolean
  limit?: number
  offset?: number
}

export function useDebtsQuery(filter: DebtFilterInput = {}) {
  return useQuery({
    queryKey: debtKeys.list(filter as Record<string, unknown>),
    queryFn: async () => {
      const res = await debtClient.listDebts({
        customerId: filter.customerId || '',
        status: filter.status || '',
        search: filter.search || '',
        overdueOnly: filter.overdueOnly || false,
        limit: filter.limit || 100,
        offset: filter.offset || 0,
      })
      return { debts: res.debts, total: res.total }
    },
  })
}

export function useDebtSummaryQuery(customerId = '') {
  return useQuery({
    queryKey: debtKeys.summary(customerId),
    queryFn: async () => {
      const res = await debtClient.getDebtSummary({ customerId })
      return res.summary
    },
  })
}

export function useCreateDebtMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (req: {
      customerId: string
      subscriptionId?: string
      amount: number
      description?: string
      debtDateUnix?: bigint
      dueDateUnix?: bigint
    }) => await debtClient.createDebt(req),
    onSuccess: () => {
      toast.success('Hutang berhasil dicatat')
      queryClient.invalidateQueries({ queryKey: debtKeys.all })
    },
    onError: (err: Error) => toast.error(`Gagal mencatat hutang: ${err.message}`),
  })
}

export function usePayDebtMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (req: {
      debtId: string
      amount: number
      paymentMethodId?: string
      notes?: string
    }) => await debtClient.payDebt(req),
    onSuccess: () => {
      toast.success('Pembayaran hutang dicatat')
      queryClient.invalidateQueries({ queryKey: debtKeys.all })
    },
    onError: (err: Error) => toast.error(`Gagal membayar hutang: ${err.message}`),
  })
}

export function useCancelDebtMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => await debtClient.cancelDebt({ id }),
    onSuccess: () => {
      toast.success('Hutang dibatalkan')
      queryClient.invalidateQueries({ queryKey: debtKeys.all })
    },
    onError: (err: Error) => toast.error(`Gagal membatalkan hutang: ${err.message}`),
  })
}

export async function exportDebtsXlsx(filter: DebtFilterInput = {}) {
  const res = await debtClient.exportDebts({
    format: 'xlsx',
    filter: {
      customerId: filter.customerId || '',
      status: filter.status || '',
      search: filter.search || '',
      overdueOnly: filter.overdueOnly || false,
    },
  })
  return res
}
