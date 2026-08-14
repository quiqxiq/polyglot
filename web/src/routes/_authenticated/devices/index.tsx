import { createFileRoute, redirect } from '@tanstack/react-router'
import { Devices } from '@/features/devices'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'

export const Route = createFileRoute('/_authenticated/devices/')({
  beforeLoad: () => {
    const permissions = useAuthStore.getState().auth.user?.permissions
    if (!canPermission(permissions, 'device:read')) {
      throw redirect({ to: '/403' })
    }
  },
  component: Devices,
})
