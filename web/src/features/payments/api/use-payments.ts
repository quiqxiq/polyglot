import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  SubmitManualPaymentRequest,
  ApprovePaymentSubmissionRequest,
  RejectPaymentSubmissionRequest,
  ReversePaymentRequest,
} from '@/gen/v1/billing_pb'
import { billingClient } from '@/lib/api-client'

const paymentKeys = {
  all: ['payment-submissions'] as const,
  list: (status: string) => [...paymentKeys.all, 'list', status] as const,
}

export function usePaymentSubmissionsQuery(status = '') {
  return useQuery({
    queryKey: paymentKeys.list(status),
    queryFn: async () => {
      const res = await billingClient.listPaymentSubmissions({ status })
      return res.submissions
    },
  })
}

export function useSubmitManualPaymentMutation() {
  return useMutation({
    mutationFn: async (req: {
      manualPaymentCode: string
      amount: number
      senderName: string
      senderBank: string
      proofImageUrl: string
      notes?: string
    }) =>
      await billingClient.submitManualPayment(
        new SubmitManualPaymentRequest(req),
      ),
  })
}

export function useApprovePaymentSubmissionMutation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (req: { id: string }) =>
      await billingClient.approvePaymentSubmission(
        new ApprovePaymentSubmissionRequest({ id: req.id }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: paymentKeys.all }),
  })
}

export function useRejectPaymentSubmissionMutation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (req: { id: string; reason: string }) =>
      await billingClient.rejectPaymentSubmission(
        new RejectPaymentSubmissionRequest({ id: req.id, reason: req.reason }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: paymentKeys.all }),
  })
}

export function useReversePaymentMutation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (req: { paymentId: string; reason: string }) =>
      await billingClient.reversePayment(
        new ReversePaymentRequest({ paymentId: req.paymentId, reason: req.reason }),
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: paymentKeys.all })
      qc.invalidateQueries({ queryKey: ['invoices'] })
    },
  })
}

export function useResolveManualPaymentMutation() {
  return useMutation({
    mutationFn: async (manualPaymentCode: string) =>
      await billingClient.resolveManualPayment({ manualPaymentCode }),
  })
}
