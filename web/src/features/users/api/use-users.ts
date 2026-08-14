import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  type CreateUserRequest,
  type UpdateUserRequest,
  type ResetPasswordRequest,
  type ToggleActiveRequest,
  type DeleteUserRequest,
} from '@/gen/v1/users_pb'
import { userClient } from '@/lib/api-client'
import { userKeys } from './keys'

export function useUsersQuery(search = '') {
  return useQuery({
    queryKey: userKeys.list(search),
    queryFn: async () => {
      const res = await userClient.listUsers({
        page: 1,
        pageSize: 100,
        search,
      })
      return { users: res.users, total: res.total }
    },
  })
}

export function useCreateUserMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: CreateUserRequest) => {
      return await userClient.createUser(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: userKeys.all })
    },
  })
}

export function useUpdateUserMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: UpdateUserRequest) => {
      return await userClient.updateUser(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: userKeys.all })
    },
  })
}

export function useResetPasswordMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: ResetPasswordRequest) => {
      return await userClient.resetPassword(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: userKeys.all })
    },
  })
}

export function useToggleActiveMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: ToggleActiveRequest) => {
      return await userClient.toggleActive(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: userKeys.all })
    },
  })
}

export function useDeleteUserMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: DeleteUserRequest) => {
      return await userClient.deleteUser(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: userKeys.all })
    },
  })
}
