import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  type CashierPayRequest,
  type CreateSubscriptionRequest,
  type UpdateSubscriptionRequest,
  type DeleteSubscriptionRequest,
  type ChangePlanRequest,
  type SuspendSubscriptionRequest,
  type ResumeSubscriptionRequest,
  type TerminateSubscriptionRequest,
  type ActivateSubscriptionRequest,
  type GenerateInvoicesRequest,
  ResolveMethod,
} from '@/gen/v1/billing_pb'
import { billingClient } from '@/lib/api-client'
import { billingKeys } from './keys'

export function useInvoicesQuery(customerId = '', status = '') {
  return useQuery({
    queryKey: billingKeys.invoices.list(customerId, status),
    queryFn: async () => {
      const res = await billingClient.listInvoices({
        customerId,
        status,
      })
      return res.invoices
    },
  })
}

export function useInvoiceQuery(id: string) {
  return useQuery({
    queryKey: billingKeys.invoices.detail(id),
    queryFn: async () => {
      const res = await billingClient.getInvoice({ id })
      return res.invoice
    },
    enabled: Boolean(id),
  })
}

export function useSubscriptionsQuery(customerId = '') {
  return useQuery({
    queryKey: billingKeys.subscriptions.list(customerId),
    queryFn: async () => {
      const res = await billingClient.listSubscriptions({
        customerId,
      })
      return res.subscriptions
    },
  })
}

export function useSubscriptionQuery(id: string) {
  return useQuery({
    queryKey: billingKeys.subscriptions.detail(id),
    queryFn: async () => {
      const res = await billingClient.getSubscription({ id })
      return res.subscription
    },
    enabled: Boolean(id),
  })
}

export function useCashierResolveQuery(identifier: string, method: ResolveMethod = ResolveMethod.RESOLVE_CODE, enabled = false) {
  return useQuery({
    queryKey: billingKeys.cashier.resolve(identifier, method),
    queryFn: async () => {
      return await billingClient.cashierResolve({ identifier, method })
    },
    enabled: Boolean(identifier) && enabled,
  })
}

export function useCashierPayMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: CashierPayRequest) => {
      return await billingClient.cashierPay(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: billingKeys.invoices.all() })
      queryClient.invalidateQueries({ queryKey: billingKeys.subscriptions.all() })
      queryClient.invalidateQueries({ queryKey: ['cashbook'] })
      queryClient.invalidateQueries({ queryKey: ['reports'] })
    },
  })
}

export function useChangePlanMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: ChangePlanRequest) => {
      return await billingClient.changePlan(req)
    },
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({ queryKey: billingKeys.subscriptions.all() })
      queryClient.invalidateQueries({ queryKey: billingKeys.subscriptions.detail(vars.subscriptionId) })
    },
  })
}

export function useSuspendSubscriptionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: SuspendSubscriptionRequest) => {
      return await billingClient.suspendSubscription(req)
    },
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({ queryKey: billingKeys.subscriptions.all() })
      queryClient.invalidateQueries({ queryKey: billingKeys.subscriptions.detail(vars.subscriptionId) })
    },
  })
}

export function useResumeSubscriptionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: ResumeSubscriptionRequest) => {
      return await billingClient.resumeSubscription(req)
    },
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({ queryKey: billingKeys.subscriptions.all() })
      queryClient.invalidateQueries({ queryKey: billingKeys.subscriptions.detail(vars.subscriptionId) })
    },
  })
}

export function useTerminateSubscriptionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: TerminateSubscriptionRequest) => {
      return await billingClient.terminateSubscription(req)
    },
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({ queryKey: billingKeys.subscriptions.all() })
      queryClient.invalidateQueries({ queryKey: billingKeys.subscriptions.detail(vars.subscriptionId) })
    },
  })
}

export function useActivateSubscriptionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: ActivateSubscriptionRequest) => {
      return await billingClient.activateSubscription(req)
    },
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({ queryKey: billingKeys.subscriptions.all() })
      queryClient.invalidateQueries({ queryKey: billingKeys.subscriptions.detail(vars.subscriptionId) })
    },
  })
}

export function useCreateSubscriptionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: CreateSubscriptionRequest) => {
      return await billingClient.createSubscription(req)
    },
    onSuccess: (res) => {
      queryClient.invalidateQueries({ queryKey: billingKeys.subscriptions.all() })
      if (res.subscription) {
        queryClient.invalidateQueries({
          queryKey: billingKeys.subscriptions.detail(res.subscription.id),
        })
      }
    },
  })
}

export function useUpdateSubscriptionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: UpdateSubscriptionRequest) => {
      return await billingClient.updateSubscription(req)
    },
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({ queryKey: billingKeys.subscriptions.all() })
      queryClient.invalidateQueries({
        queryKey: billingKeys.subscriptions.detail(vars.id),
      })
    },
  })
}

export function useDeleteSubscriptionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: DeleteSubscriptionRequest) => {
      return await billingClient.deleteSubscription(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: billingKeys.subscriptions.all() })
    },
  })
}

export function useGenerateInvoicesMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: GenerateInvoicesRequest) => {
      return await billingClient.generateInvoices(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: billingKeys.invoices.all() })
      queryClient.invalidateQueries({ queryKey: ['reports'] })
    },
  })
}
