import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  type RequestOTPRequest,
  type PortalLoginRequest,
  type PortalLogoutRequest,
} from '@/gen/v1/portal_pb'
import { portalClient } from '@/lib/api-client'
import { portalKeys } from './keys'

export function usePortalOverviewQuery(token: string) {
  return useQuery({
    queryKey: portalKeys.overview(token),
    queryFn: async () => {
      return await portalClient.overview({ token })
    },
    enabled: Boolean(token),
  })
}

export function usePortalInvoicesQuery(token: string, limit = 50) {
  return useQuery({
    queryKey: portalKeys.invoices(token, limit),
    queryFn: async () => {
      const res = await portalClient.myInvoices({ token, limit })
      return res.invoices
    },
    enabled: Boolean(token),
  })
}

export function usePortalPaymentsQuery(token: string, limit = 50) {
  return useQuery({
    queryKey: portalKeys.payments(token, limit),
    queryFn: async () => {
      const res = await portalClient.myPayments({ token, limit })
      return res.payments
    },
    enabled: Boolean(token),
  })
}

export function useRequestOTPMutation() {
  return useMutation({
    mutationFn: async (req: RequestOTPRequest) => {
      return await portalClient.requestOTP(req)
    },
  })
}

export function usePortalLoginMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: PortalLoginRequest) => {
      return await portalClient.login(req)
    },
    onSuccess: (res) => {
      if (res.token) {
        queryClient.invalidateQueries({ queryKey: portalKeys.overview(res.token) })
      }
    },
  })
}

export function usePortalLogoutMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: PortalLogoutRequest) => {
      return await portalClient.logout(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: portalKeys.all })
    },
  })
}
