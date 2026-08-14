import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  AddPolicyRequest,
  AssignRoleRequest,
  RemovePolicyRequest,
  UnassignRoleRequest,
  type Policy,
  type RoleAssignment,
} from '@/gen/v1/rbac_pb'
import { rbacClient } from '@/lib/api-client'
import { rbacKeys } from './keys'

export function usePoliciesQuery() {
  return useQuery({
    queryKey: rbacKeys.policies(),
    queryFn: async () => {
      const res = await rbacClient.listPolicies({})
      return res.policies
    },
  })
}

export function useRoleAssignmentsQuery() {
  return useQuery({
    queryKey: rbacKeys.assignments(),
    queryFn: async () => {
      const res = await rbacClient.listRoleAssignments({})
      return res.roleAssignments
    },
  })
}

export function useAddPolicyMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (policy: Policy) => {
      return await rbacClient.addPolicy(new AddPolicyRequest({ policy }))
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: rbacKeys.policies() })
    },
  })
}

export function useRemovePolicyMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (policy: Policy) => {
      return await rbacClient.removePolicy(new RemovePolicyRequest({ policy }))
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: rbacKeys.policies() })
    },
  })
}

export function useAssignRoleMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (assignment: RoleAssignment) => {
      return await rbacClient.assignRole(new AssignRoleRequest({ assignment }))
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: rbacKeys.assignments() })
    },
  })
}

export function useUnassignRoleMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (assignment: RoleAssignment) => {
      return await rbacClient.unassignRole(
        new UnassignRoleRequest({ assignment })
      )
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: rbacKeys.assignments() })
    },
  })
}
