import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  type CreatePlanRequest,
  type UpdatePlanRequest,
  type DeletePlanRequest,
} from '@/gen/v1/billing_pb'
import { billingClient } from '@/lib/api-client'
import { billingKeys } from './keys'

export function usePlansQuery(activeOnly = false) {
  return useQuery({
    queryKey: billingKeys.plans.list(activeOnly),
    queryFn: async () => {
      const res = await billingClient.listPlans({
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
      const res = await billingClient.getPlan({ id })
      return res.plan
    },
    enabled: Boolean(id),
  })
}

export function useCreatePlanMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: CreatePlanRequest) => {
      return await billingClient.createPlan(req)
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
      return await billingClient.updatePlan(req)
    },
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({ queryKey: billingKeys.plans.all() })
      if (vars.plan?.id) {
        queryClient.invalidateQueries({ queryKey: billingKeys.plans.detail(vars.plan.id) })
      }
    },
  })
}

export function useDeletePlanMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: DeletePlanRequest) => {
      return await billingClient.deletePlan(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: billingKeys.plans.all() })
    },
  })
}
