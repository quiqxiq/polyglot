import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { customerClient } from '@/lib/api-client'
import { customerKeys } from './keys'
import {
  CreateCustomerRequest,
  UpdateCustomerRequest,
  DeleteCustomerRequest,
} from '@/gen/v1/customer_pb'

export function useCustomersQuery() {
  return useQuery({
    queryKey: customerKeys.lists(),
    queryFn: async () => {
      const res = await customerClient.listCustomers({})
      return res.customers
    },
  })
}

export function useCustomerQuery(id: string, enabled = true) {
  return useQuery({
    queryKey: customerKeys.detail(id),
    queryFn: async () => {
      const res = await customerClient.getCustomer({ id })
      return res.customer
    },
    enabled: Boolean(id) && enabled,
  })
}

export function useCreateCustomerMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: CreateCustomerRequest) => {
      return await customerClient.createCustomer(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: customerKeys.lists() })
    },
  })
}

export function useUpdateCustomerMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: UpdateCustomerRequest) => {
      return await customerClient.updateCustomer(req)
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: customerKeys.lists() })
      if (variables.customer?.id) {
        queryClient.invalidateQueries({ queryKey: customerKeys.detail(variables.customer.id) })
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
      queryClient.invalidateQueries({ queryKey: customerKeys.lists() })
    },
  })
}
