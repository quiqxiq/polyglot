import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  type CreateCustomerRequest,
  type UpdateCustomerRequest,
  type DeleteCustomerRequest,
} from '@/gen/v1/customer_pb'
import {
  type ImportFileRequest,
  type ImportRouterRequest,
  type ExportCustomersRequest,
} from '@/gen/v1/ispadmin_pb'
import { customerClient, ispAdminClient } from '@/lib/api-client'
import { customerKeys } from './keys'

// ─── Customer CRM Queries & Mutations ───────────────────────────────────

export function useCustomersQuery() {
  return useQuery({
    queryKey: customerKeys.list(),
    queryFn: async () => {
      const res = await customerClient.listCustomers({})
      return res.customers
    },
  })
}

export function useCustomerQuery(id: string) {
  return useQuery({
    queryKey: customerKeys.detail(id),
    queryFn: async () => {
      const res = await customerClient.getCustomer({ id })
      return res.customer
    },
    enabled: Boolean(id),
  })
}

export function useFindCustomerByPhoneQuery(phone: string, enabled = false) {
  return useQuery({
    queryKey: customerKeys.lookup('phone', phone),
    queryFn: async () => {
      const res = await customerClient.findByPhone({ phone })
      return res.customer
    },
    enabled: Boolean(phone) && enabled,
  })
}

export function useFindCustomerByCodeQuery(customerCode: string, enabled = false) {
  return useQuery({
    queryKey: customerKeys.lookup('code', customerCode),
    queryFn: async () => {
      const res = await customerClient.findByCustomerCode({ customerCode })
      return res.customer
    },
    enabled: Boolean(customerCode) && enabled,
  })
}

export function useFindCustomerByPortalCodeQuery(portalAccessCode: string, enabled = false) {
  return useQuery({
    queryKey: customerKeys.lookup('portal', portalAccessCode),
    queryFn: async () => {
      const res = await customerClient.findByPortalCode({ portalAccessCode })
      return res.customer
    },
    enabled: Boolean(portalAccessCode) && enabled,
  })
}

export function useCreateCustomerMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: CreateCustomerRequest) => {
      return await customerClient.createCustomer(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: customerKeys.all })
    },
  })
}

export function useUpdateCustomerMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: UpdateCustomerRequest) => {
      return await customerClient.updateCustomer(req)
    },
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({ queryKey: customerKeys.all })
      if (vars.customer?.id) {
        queryClient.invalidateQueries({ queryKey: customerKeys.detail(vars.customer.id) })
      }
    },
  })
}

export function useDeleteCustomerMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: DeleteCustomerRequest) => {
      return await customerClient.deleteCustomer(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: customerKeys.all })
    },
  })
}

// ─── Import, Export & Router Drift Reconcile ────────────────────────────

export function useReconcileQuery(deviceId: string, enabled = false) {
  return useQuery({
    queryKey: customerKeys.reconcile(deviceId),
    queryFn: async () => {
      return await ispAdminClient.reconcile({ deviceId })
    },
    enabled: Boolean(deviceId) && enabled,
  })
}

export function useImportFileMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: ImportFileRequest) => {
      return await ispAdminClient.importFile(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: customerKeys.all })
      queryClient.invalidateQueries({ queryKey: ['billing'] })
    },
  })
}

export function useImportRouterMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: ImportRouterRequest) => {
      return await ispAdminClient.importRouter(req)
    },
    onSuccess: (_, vars) => {
      if (!vars.dryRun) {
        queryClient.invalidateQueries({ queryKey: customerKeys.all })
        queryClient.invalidateQueries({ queryKey: ['billing'] })
      }
    },
  })
}

export function useExportCustomersMutation() {
  return useMutation({
    mutationFn: async (req: ExportCustomersRequest) => {
      return await ispAdminClient.exportCustomers(req)
    },
  })
}
