import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { RBACPolicy, RBACRoleAssignment } from '../types'
import {
  listPoliciesApi,
  addPolicyApi,
  removePolicyApi,
  listRoleAssignmentsApi,
  assignRoleApi,
  unassignRoleApi,
} from '../api/client'

export const useRBACStore = defineStore('rbac', () => {
  const policies = ref<RBACPolicy[]>([])
  const roleAssignments = ref<RBACRoleAssignment[]>([])
  const loading = ref<boolean>(false)
  const error = ref<string | null>(null)

  async function fetchPolicies() {
    try {
      loading.value = true
      error.value = null
      const res = await listPoliciesApi()
      policies.value = (res.policies || []).map(([role, path, method]) => ({
        role,
        path,
        method,
      }))
    } catch (e: any) {
      error.value = e.message || 'Failed to fetch RBAC policies'
    } finally {
      loading.value = false
    }
  }

  async function addPolicy(role: string, path: string, method: string) {
    try {
      loading.value = true
      error.value = null
      await addPolicyApi(role, path, method)
      policies.value.push({ role, path, method })
    } catch (e: any) {
      error.value = e.message || 'Failed to add policy'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function removePolicy(role: string, path: string, method: string) {
    try {
      loading.value = true
      error.value = null
      await removePolicyApi(role, path, method)
      policies.value = policies.value.filter(
        (p) => !(p.role === role && p.path === path && p.method === method)
      )
    } catch (e: any) {
      error.value = e.message || 'Failed to remove policy'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function fetchRoleAssignments() {
    try {
      loading.value = true
      error.value = null
      const res = await listRoleAssignmentsApi()
      roleAssignments.value = (res.roles || []).map(([user, role]) => ({
        user,
        role,
      }))
    } catch (e: any) {
      error.value = e.message || 'Failed to fetch role assignments'
    } finally {
      loading.value = false
    }
  }

  async function assignRole(user: string, role: string) {
    try {
      loading.value = true
      error.value = null
      await assignRoleApi(user, role)
      roleAssignments.value.push({ user, role })
    } catch (e: any) {
      error.value = e.message || 'Failed to assign role'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function unassignRole(user: string, role: string) {
    try {
      loading.value = true
      error.value = null
      await unassignRoleApi(user, role)
      roleAssignments.value = roleAssignments.value.filter(
        (r) => !(r.user === user && r.role === role)
      )
    } catch (e: any) {
      error.value = e.message || 'Failed to unassign role'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function fetchAll() {
    await Promise.allSettled([fetchPolicies(), fetchRoleAssignments()])
  }

  return {
    policies,
    roleAssignments,
    loading,
    error,
    fetchPolicies,
    addPolicy,
    removePolicy,
    fetchRoleAssignments,
    assignRole,
    unassignRole,
    fetchAll,
  }
})
