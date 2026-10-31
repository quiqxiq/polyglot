import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  type CashierPayRequest,
  type GenerateInvoicesRequest,
  type ResolveMethod,
  ApplyInvoiceDiscountRequest,
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
  RetryProvisionRequest,
} from '@/gen/v1/subscription_pb'
import { useAuthStore } from '@/stores/auth-store'
import { API_BASE_URL } from '@/lib/api-base'
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
      queryClient.invalidateQueries({
        queryKey: billingKeys.subscriptions.all(),
      })
    },
  })
}

export interface CashierChargeRequest {
  invoice_id: string
  gateway?: string
  channel?: string
  expire_minutes?: number
}

export interface PaymentGatewayOption {
  name: string
  enabled: boolean
  is_default: boolean
}

// useGatewaysQuery mengambil daftar payment gateway terdaftar + status aktif
// untuk pemilih gateway dinamis pada kasir (planning v2 Fase 3).
export function useGatewaysQuery() {
  return useQuery({
    queryKey: ['payment-gateways'],
    queryFn: async (): Promise<PaymentGatewayOption[]> => {
      const baseUrl = API_BASE_URL
      const token = useAuthStore.getState().auth.accessToken
      const res = await fetch(`${baseUrl}/api/cashier/gateways`, {
        headers: {
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
      })
      if (!res.ok) return []
      const data = (await res.json()) as { gateways?: PaymentGatewayOption[] }
      return (data.gateways || []).filter((g) => g.enabled)
    },
    staleTime: 60_000,
  })
}

export interface CashierChargeResponse {
  external_id: string
  payment_url: string
  qr_string: string
  va_number: string
  status: string
  amount: number
}

export function useCashierChargeMutation() {
  return useMutation({
    mutationFn: async (
      req: CashierChargeRequest
    ): Promise<CashierChargeResponse> => {
      const baseUrl = API_BASE_URL
      const token = useAuthStore.getState().auth.accessToken
      const res = await fetch(`${baseUrl}/api/cashier/charge`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: JSON.stringify(req),
      })
      if (!res.ok) {
        const data = await res.json().catch(() => null)
        throw new Error(data?.message || 'Gagal membuat tagihan online')
      }
      return res.json()
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
      queryClient.invalidateQueries({
        queryKey: billingKeys.subscriptions.all(),
      })
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
      queryClient.invalidateQueries({
        queryKey: billingKeys.subscriptions.all(),
      })
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
      queryClient.invalidateQueries({
        queryKey: billingKeys.subscriptions.all(),
      })
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
      queryClient.invalidateQueries({
        queryKey: billingKeys.subscriptions.all(),
      })
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
      queryClient.invalidateQueries({
        queryKey: billingKeys.subscriptions.all(),
      })
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
      queryClient.invalidateQueries({
        queryKey: billingKeys.subscriptions.all(),
      })
      queryClient.invalidateQueries({
        queryKey: billingKeys.subscriptions.detail(vars.subscriptionId),
      })
    },
  })
}

export function useRetryProvisionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: { id: string }) => {
      return await subscriptionClient.retryProvision(
        new RetryProvisionRequest({ id: req.id })
      )
    },
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({
        queryKey: billingKeys.subscriptions.all(),
      })
      queryClient.invalidateQueries({
        queryKey: billingKeys.subscriptions.detail(vars.id),
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
      queryClient.invalidateQueries({
        queryKey: billingKeys.subscriptions.all(),
      })
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
      queryClient.invalidateQueries({
        queryKey: billingKeys.subscriptions.all(),
      })
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
      queryClient.invalidateQueries({
        queryKey: billingKeys.subscriptions.all(),
      })
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
      queryClient.invalidateQueries({
        queryKey: billingKeys.subscriptions.all(),
      })
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

export function useApplyInvoiceDiscountMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: {
      invoiceId: string
      discountType: string
      discountValue: number
      promoCode?: string
    }) =>
      await billingClient.applyInvoiceDiscount(
        new ApplyInvoiceDiscountRequest({
          invoiceId: req.invoiceId,
          discountType: req.discountType,
          discountValue: req.discountValue,
          promoCode: req.promoCode || '',
        })
      ),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: billingKeys.invoices.all() })
    },
  })
}
