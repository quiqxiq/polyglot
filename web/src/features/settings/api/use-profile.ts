import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { authClient } from '@/lib/api-client'
import { useAuthStore } from '@/stores/auth-store'
import { toast } from 'sonner'

export function useGetMe() {
  return useQuery({
    queryKey: ['auth', 'me'],
    queryFn: async () => {
      const res = await authClient.getMe({})
      if (res.user) {
        useAuthStore.getState().auth.setUser({
          accountNo: res.user.id,
          username: res.user.username,
          fullName: res.user.fullName,
          phoneNumber: res.user.phoneNumber,
          specialization: res.user.specialization,
          email: res.user.email,
          role: res.user.roles.length ? res.user.roles : [res.user.role],
          permissions: res.user.permissions,
        })
      }
      return res.user
    },
  })
}

export function useUpdateMe() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: {
      fullName: string
      phoneNumber: string
      email: string
      specialization: string
    }) => {
      const res = await authClient.updateMe({
        fullName: data.fullName,
        phoneNumber: data.phoneNumber,
        email: data.email,
        specialization: data.specialization,
      })
      return res.user
    },
    onSuccess: (updatedUser) => {
      if (updatedUser) {
        useAuthStore.getState().auth.setUser({
          accountNo: updatedUser.id,
          username: updatedUser.username,
          fullName: updatedUser.fullName,
          phoneNumber: updatedUser.phoneNumber,
          specialization: updatedUser.specialization,
          email: updatedUser.email,
          role: updatedUser.roles.length ? updatedUser.roles : [updatedUser.role],
          permissions: updatedUser.permissions,
        })
      }
      queryClient.invalidateQueries({ queryKey: ['auth', 'me'] })
      toast.success('Profil berhasil diperbarui!')
    },
    onError: (err: Error) => {
      toast.error(`Gagal memperbarui profil: ${err.message}`)
    },
  })
}

export function useChangePassword() {
  return useMutation({
    mutationFn: async (data: { oldPassword: string; newPassword: string }) => {
      return await authClient.changePassword({
        oldPassword: data.oldPassword,
        newPassword: data.newPassword,
      })
    },
    onSuccess: () => {
      toast.success('Password berhasil diubah!')
    },
    onError: (err: Error) => {
      toast.error(`Gagal mengubah password: ${err.message}`)
    },
  })
}
