import { createFileRoute, redirect } from '@tanstack/react-router'
import { useAuthStore } from '@/stores/auth-store'
import { authClient } from '@/lib/api-client'
import { AuthenticatedLayout } from '@/components/layout/authenticated-layout'

async function trySilentRefresh(): Promise<boolean> {
  const { setAccessToken, setUser } = useAuthStore.getState().auth
  try {
    const res = await authClient.refreshToken({})
    if (!res.token) return false
    setAccessToken(res.token)
    // Isi identitas user dari GetMe supaya sidebar/menu punya roles terbaru.
    const me = await authClient.getMe({})
    if (me.user) {
      setUser({
        accountNo: me.user.id,
        email: me.user.email,
        role: me.user.roles.length ? me.user.roles : [me.user.role],
        exp: Number(res.expiresAtUnix) * 1000,
      })
    }
    return true
  } catch {
    return false
  }
}

export const Route = createFileRoute('/_authenticated')({
  beforeLoad: async ({ location }) => {
    const { accessToken, user, reset } = useAuthStore.getState().auth
    const isExpired = user?.exp ? user.exp < Date.now() : false

    if (!accessToken || isExpired) {
      // Access token memory-only hilang saat reload — coba refresh lewat
      // cookie httpOnly sebelum mengusir user ke sign-in.
      const refreshed = await trySilentRefresh()
      if (!refreshed) {
        reset()
        throw redirect({
          to: '/sign-in',
          search: {
            redirect: location.href,
          },
        })
      }
    }
  },
  component: AuthenticatedLayout,
})
