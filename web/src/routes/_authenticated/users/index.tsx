import { createFileRoute, redirect } from '@tanstack/react-router'
import { Users } from '@/features/users'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'

export const Route = createFileRoute('/_authenticated/users/')({
  beforeLoad: () => {
    const permissions = useAuthStore.getState().auth.user?.permissions
    if (!canPermission(permissions, 'user:read')) {
      throw redirect({ to: '/403' })
    }
  },
  component: Users,
})
