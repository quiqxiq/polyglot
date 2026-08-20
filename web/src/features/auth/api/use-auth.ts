import { useMutation, useQuery } from '@tanstack/react-query'
import { type LoginRequest, type RefreshTokenRequest } from '@/gen/v1/auth_pb'
import { useAuthStore } from '@/stores/auth-store'
import { authClient } from '@/lib/api-client'
import { authKeys } from './keys'

export function useLoginMutation() {
  const setAccessToken = useAuthStore((s) => s.auth.setAccessToken)

  return useMutation({
    mutationFn: async (req: LoginRequest) => {
      const res = await authClient.login(req)
      if (res.token) {
        setAccessToken(res.token)
      }
      return res
    },
  })
}

export function useMeQuery(enabled = true) {
  return useQuery({
    queryKey: authKeys.me(),
    queryFn: async () => {
      const res = await authClient.getMe({})
      return res.user
    },
    enabled,
  })
}

export function useRefreshTokenMutation() {
  const setAccessToken = useAuthStore((s) => s.auth.setAccessToken)

  return useMutation({
    mutationFn: async (req: RefreshTokenRequest) => {
      const res = await authClient.refreshToken(req)
      if (res.token) {
        setAccessToken(res.token)
      }
      return res
    },
  })
}

// Logout: revoke refresh token di server (cookie httpOnly dibersihkan via
// Set-Cookie expired), lalu reset state lokal.
export function useLogoutMutation() {
  const reset = useAuthStore((s) => s.auth.reset)

  return useMutation({
    mutationFn: async () => {
      const res = await authClient.logout({})
      reset()
      return res
    },
  })
}
