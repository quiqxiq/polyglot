import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  type CashAccount,
  type CashCategory,
  type AddTransactionRequest,
} from '@/gen/v1/cashbook_pb'
import { cashbookClient } from '@/lib/api-client'
import { cashbookKeys } from './keys'

export function useCashAccountsQuery(activeOnly = false) {
  return useQuery({
    queryKey: cashbookKeys.accounts.list(activeOnly),
    queryFn: async () => {
      const res = await cashbookClient.listAccounts({ activeOnly })
      return res.accounts
    },
  })
}

export function useCashCategoriesQuery(activeOnly = false) {
  return useQuery({
    queryKey: cashbookKeys.categories.list(activeOnly),
    queryFn: async () => {
      const res = await cashbookClient.listCategories({ activeOnly })
      return res.categories
    },
  })
}

export interface TransactionFilterParams {
  accountId?: string
  categoryId?: string
  direction?: string
  fromUnix?: number
  toUnix?: number
  limit?: number
}

export function useCashTransactionsQuery(filter: TransactionFilterParams = {}) {
  return useQuery({
    queryKey: cashbookKeys.transactions.list(filter),
    queryFn: async () => {
      const res = await cashbookClient.listTransactions({
        accountId: filter.accountId || '',
        categoryId: filter.categoryId || '',
        direction: filter.direction || '',
        fromUnix: filter.fromUnix ? BigInt(filter.fromUnix) : BigInt(0),
        toUnix: filter.toUnix ? BigInt(filter.toUnix) : BigInt(0),
        limit: filter.limit || 50,
      })
      return res.transactions
    },
  })
}

export function useCashBalancesQuery(fromUnix?: number, toUnix?: number) {
  return useQuery({
    queryKey: cashbookKeys.balances(fromUnix, toUnix),
    queryFn: async () => {
      const res = await cashbookClient.balances({
        fromUnix: fromUnix ? BigInt(fromUnix) : BigInt(0),
        toUnix: toUnix ? BigInt(toUnix) : BigInt(0),
      })
      return res.balanceByAccount
    },
  })
}

export function useSaveCashAccountMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (account: Partial<CashAccount>) => {
      return await cashbookClient.saveAccount({ account: account as CashAccount })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: cashbookKeys.accounts.all() })
      queryClient.invalidateQueries({ queryKey: cashbookKeys.all })
    },
  })
}

export function useSaveCashCategoryMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (category: Partial<CashCategory>) => {
      return await cashbookClient.saveCategory({ category: category as CashCategory })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: cashbookKeys.categories.all() })
      queryClient.invalidateQueries({ queryKey: cashbookKeys.all })
    },
  })
}

export function useAddCashTransactionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: AddTransactionRequest) => {
      return await cashbookClient.addTransaction(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: cashbookKeys.transactions.all() })
      queryClient.invalidateQueries({ queryKey: cashbookKeys.all })
      queryClient.invalidateQueries({ queryKey: ['reports'] })
    },
  })
}
