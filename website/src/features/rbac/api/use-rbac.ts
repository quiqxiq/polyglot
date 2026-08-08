import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { rbacClient } from '@/lib/api-client'
import { rbacKeys } from './keys'
import {
  AddPolicyRequest,
  RemovePolicyRequest,
  AssignRoleRequest,
  UnassignRoleRequest,
} from '@/gen/v1/rbac_pb'

export function usePoliciesQuery() {
  return useQuery({
    queryKey: rbacKeys.policies(),
    queryFn: async () => {
      const res = await rbacClient.listPolicies({})
      return res.policies
    },
  })
}

export function useAddPolicyMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: AddPolicyRequest) => {
      return await rbacClient.addPolicy(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: rbacKeys.policies() })
    },
  })
}

export function useRemovePolicyMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: RemovePolicyRequest) => {
      return await rbacClient.removePolicy(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: rbacKeys.policies() })
    },
  })
}

export function useRoleAssignmentsQuery() {
  return useQuery({
    queryKey: rbacKeys.roleAssignments(),
    queryFn: async () => {
      const res = await rbacClient.listRoleAssignments({})
      return res.roleAssignments
    },
  })
}

export function useAssignRoleMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: AssignRoleRequest) => {
      return await rbacClient.assignRole(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: rbacKeys.roleAssignments() })
    },
  })
}

export function useUnassignRoleMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: UnassignRoleRequest) => {
      return await rbacClient.unassignRole(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: rbacKeys.roleAssignments() })
    },
  })
}
