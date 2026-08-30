import { createFileRoute, redirect } from '@tanstack/react-router'
import { Registrations } from '@/features/registration'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'

export const Route = createFileRoute('/_authenticated/registrations/')({
  beforeLoad: () => {
    const permissions = useAuthStore.getState().auth.user?.permissions
    if (!canPermission(permissions, 'customer:read')) {
      throw redirect({ to: '/403' })
    }
  },
  component: Registrations,
})
