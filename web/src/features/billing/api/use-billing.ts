import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { billingClient } from '@/lib/api-client'
import { billingKeys } from './keys'
import {
  CreateInvoiceRequest,
  PayInvoiceRequest,
  CreateSubscriptionRequest,
  CancelSubscriptionRequest,
} from '@/gen/v1/billing_pb'

export function useInvoicesQuery(customerId = '') {
  return useQuery({
    queryKey: billingKeys.invoices(customerId),
    queryFn: async () => {
      const res = await billingClient.listInvoices({ customerId })
      return res.invoices
    },
  })
}

export function useCreateInvoiceMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: CreateInvoiceRequest) => {
      return await billingClient.createInvoice(req)
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: billingKeys.invoices(variables.customerId) })
    },
  })
}

export function usePayInvoiceMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: PayInvoiceRequest) => {
      return await billingClient.payInvoice(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: billingKeys.invoices() })
    },
  })
}

export function useSubscriptionsQuery(customerId = '') {
  return useQuery({
    queryKey: billingKeys.subscriptions(customerId),
    queryFn: async () => {
      const res = await billingClient.listSubscriptions({ customerId })
      return res.subscriptions
    },
  })
}

export function useCreateSubscriptionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: CreateSubscriptionRequest) => {
      return await billingClient.createSubscription(req)
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: billingKeys.subscriptions(variables.customerId) })
    },
  })
}

export function useCancelSubscriptionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: CancelSubscriptionRequest) => {
      return await billingClient.cancelSubscription(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: billingKeys.subscriptions() })
    },
  })
}
