import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  ApprovalDecision,
  type DecideApprovalRequest,
} from '@/gen/v1/approval_pb'
import { toast } from 'sonner'
import { approvalClient } from '@/lib/api-client'
import { aiOpsKeys } from './keys'

export interface ApprovalFilterInput {
  deviceId?: string
  status?: string
  limit?: number
  offset?: number
}

export function useApprovalsQuery(filter: ApprovalFilterInput = {}) {
  return useQuery({
    queryKey: aiOpsKeys.approvals.list(filter as Record<string, unknown>),
    queryFn: async () => {
      const res = await approvalClient.listApprovals({
        deviceId: filter.deviceId || '',
        status: filter.status || '',
        limit: filter.limit || 50,
        offset: filter.offset || 0,
      })
      return { approvals: res.approvals, total: res.total }
    },
    refetchInterval: 10_000,
  })
}

export function useApprovalQuery(id: string) {
  return useQuery({
    queryKey: aiOpsKeys.approvals.detail(id),
    queryFn: async () => (await approvalClient.getApproval({ id })).approval,
    enabled: Boolean(id),
    refetchInterval: 10_000,
  })
}

export function useDecideApprovalMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (req: Partial<DecideApprovalRequest>) =>
      await approvalClient.decideApproval(req as DecideApprovalRequest),
    onSuccess: (res) => {
      const status = res.approval?.status
      toast.success(`Approval diproses: ${status || 'OK'}`)
      queryClient.invalidateQueries({ queryKey: aiOpsKeys.approvals.all() })
      queryClient.invalidateQueries({ queryKey: aiOpsKeys.restorePoints.all() })
    },
    onError: (err: Error) =>
      toast.error(`Gagal memproses approval: ${err.message}`),
  })
}

export function useCancelApprovalMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) =>
      await approvalClient.cancelApproval({ id }),
    onSuccess: () => {
      toast.success('Approval dibatalkan')
      queryClient.invalidateQueries({ queryKey: aiOpsKeys.approvals.all() })
    },
    onError: (err: Error) => toast.error(`Gagal membatalkan: ${err.message}`),
  })
}

export { ApprovalDecision }
