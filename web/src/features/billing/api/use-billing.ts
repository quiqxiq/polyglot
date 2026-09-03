import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  type CashierPayRequest,
  type GenerateInvoicesRequest,
  ResolveMethod,
} from '@/gen/v1/billing_pb'
import {
  type CreateSubscriptionRequest,
  type UpdateSubscriptionRequest,
  type DeleteSubscriptionRequest,
  type ChangePlanRequest,
  type SuspendSubscriptionRequest,
  type ResumeSubscriptionRequest,
  type TerminateSubscriptionRequest,
  type ActivateSubscriptionRequest,
  type IsolateSubscriptionRequest,
  type RestoreSubscriptionRequest,
} from '@/gen/v1/subscription_pb'
import { billingClient, subscriptionClient } from '@/lib/api-client'
import { billingKeys } from './keys'

export function useInvoicesQuery(
  customerId = '',
  status = '',
  options?: { enabled?: boolean }
) {
  return useQuery({
    queryKey: billingKeys.invoices.list(customerId, status),
    queryFn: async () => {
      const res = await billingClient.listInvoices({
        customerId,
        status,
      })
      return res.invoices
    },
    enabled: options?.enabled,
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

export function useSubscriptionsQuery(
  customerId = '',
  options?: { enabled?: boolean }
) {
  return useQuery({
    queryKey: billingKeys.subscriptions.list(customerId),
    queryFn: async () => {
      const res = await subscriptionClient.listSubscriptions({
        customerId,
      })
      return res.subscriptions
    },
    enabled: options?.enabled,
  })
}

export function useSubscriptionQuery(id: string) {
  return useQuery({
    queryKey: billingKeys.subscriptions.detail(id),
    queryFn: async () => {
      const res = await subscriptionClient.getSubscription({ id })
      return res.subscription
    },
    enabled: Boolean(id),
  })
}

export function useCashierResolveQuery(
  identifier: string,
  method: ResolveMethod,
  enabled = true
) {
  return useQuery({
    queryKey: billingKeys.cashier.resolve(identifier, method),
    queryFn: async () => {
      if (!identifier) return null
      return await billingClient.cashierResolve({ identifier, method })
    },
    enabled: Boolean(identifier) && enabled,
  })
}

export function useCashierResolveMutation() {
  return useMutation({
    mutationFn: async ({
      identifier,
      method,
    }: {
      identifier: string
      method: ResolveMethod
    }) => {
      return await billingClient.cashierResolve({ identifier, method })
    },
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
    },
  })
}

export function useChangePlanMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: ChangePlanRequest) => {
      return await subscriptionClient.changePlan(req)
    },
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({ queryKey: billingKeys.subscriptions.all() })
      queryClient.invalidateQueries({
        queryKey: billingKeys.subscriptions.detail(vars.subscriptionId),
      })
    },
  })
}

export function useSuspendSubscriptionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: SuspendSubscriptionRequest) => {
      return await subscriptionClient.suspendSubscription(req)
    },
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({ queryKey: billingKeys.subscriptions.all() })
      queryClient.invalidateQueries({
        queryKey: billingKeys.subscriptions.detail(vars.subscriptionId),
      })
    },
  })
}

export function useResumeSubscriptionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: ResumeSubscriptionRequest) => {
      return await subscriptionClient.resumeSubscription(req)
    },
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({ queryKey: billingKeys.subscriptions.all() })
      queryClient.invalidateQueries({
        queryKey: billingKeys.subscriptions.detail(vars.subscriptionId),
      })
    },
  })
}

export function useTerminateSubscriptionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: TerminateSubscriptionRequest) => {
      return await subscriptionClient.terminateSubscription(req)
    },
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({ queryKey: billingKeys.subscriptions.all() })
      queryClient.invalidateQueries({
        queryKey: billingKeys.subscriptions.detail(vars.subscriptionId),
      })
    },
  })
}

export function useActivateSubscriptionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: ActivateSubscriptionRequest) => {
      return await subscriptionClient.activateSubscription(req)
    },
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({ queryKey: billingKeys.subscriptions.all() })
      queryClient.invalidateQueries({
        queryKey: billingKeys.subscriptions.detail(vars.subscriptionId),
      })
    },
  })
}

export function useIsolateSubscriptionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: IsolateSubscriptionRequest) => {
      return await subscriptionClient.isolateSubscription(req)
    },
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({ queryKey: billingKeys.subscriptions.all() })
      queryClient.invalidateQueries({
        queryKey: billingKeys.subscriptions.detail(vars.subscriptionId),
      })
    },
  })
}

export function useRestoreSubscriptionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: RestoreSubscriptionRequest) => {
      return await subscriptionClient.restoreSubscription(req)
    },
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({ queryKey: billingKeys.subscriptions.all() })
      queryClient.invalidateQueries({
        queryKey: billingKeys.subscriptions.detail(vars.subscriptionId),
      })
    },
  })
}

export function useCreateSubscriptionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: CreateSubscriptionRequest) => {
      return await subscriptionClient.createSubscription(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: billingKeys.subscriptions.all() })
    },
  })
}

export function useUpdateSubscriptionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: UpdateSubscriptionRequest) => {
      return await subscriptionClient.updateSubscription(req)
    },
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({ queryKey: billingKeys.subscriptions.all() })
      if (vars.id) {
        queryClient.invalidateQueries({
          queryKey: billingKeys.subscriptions.detail(vars.id),
        })
      }
    },
  })
}

export function useDeleteSubscriptionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: DeleteSubscriptionRequest) => {
      return await subscriptionClient.deleteSubscription(req)
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
    },
  })
}
