import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  type PullRouterPlansRequest,
  type CommitPlansRequest,
} from '@/gen/v1/ispadmin_pb'
import {
  type CreatePlanRequest,
  type UpdatePlanRequest,
  type DeletePlanRequest,
} from '@/gen/v1/plan_pb'
import { planClient, ispAdminClient } from '@/lib/api-client'
import { billingKeys } from './keys'

export function usePlansQuery(activeOnly = false) {
  return useQuery({
    queryKey: billingKeys.plans.list(activeOnly),
    queryFn: async () => {
      const res = await planClient.listPlans({
        activeOnly,
      })
      return res.plans
    },
  })
}

export function usePlanQuery(id: string) {
  return useQuery({
    queryKey: billingKeys.plans.detail(id),
    queryFn: async () => {
      const res = await planClient.getPlan({ id })
      return res.plan
    },
    enabled: Boolean(id),
  })
}

export function useCreatePlanMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: CreatePlanRequest) => {
      return await planClient.createPlan(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: billingKeys.plans.all() })
    },
  })
}

export function useUpdatePlanMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: UpdatePlanRequest) => {
      return await planClient.updatePlan(req)
    },
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({ queryKey: billingKeys.plans.all() })
      if (vars.plan?.id) {
        queryClient.invalidateQueries({
          queryKey: billingKeys.plans.detail(vars.plan.id),
        })
      }
    },
  })
}

export function useDeletePlanMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: DeletePlanRequest) => {
      return await planClient.deletePlan(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: billingKeys.plans.all() })
    },
  })
}

export function usePullRouterPlansMutation() {
  return useMutation({
    mutationFn: async (req: PullRouterPlansRequest) => {
      return await ispAdminClient.pullRouterPlans(req)
    },
  })
}

export function useCommitPlansMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: CommitPlansRequest) => {
      return await ispAdminClient.commitPlans(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: billingKeys.plans.all() })
    },
  })
}
