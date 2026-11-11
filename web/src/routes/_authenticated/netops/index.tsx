import { createFileRoute, redirect } from '@tanstack/react-router'
import { canPermission } from '@/hooks/use-can'
import { useAuthStore } from '@/stores/auth-store'
import { NetOpsLaunch } from '@/features/netops'

export const Route = createFileRoute('/_authenticated/netops/')({
  beforeLoad: () => {
    const permissions = useAuthStore.getState().auth.user?.permissions
    if (!canPermission(permissions, 'harness:admin')) {
      throw redirect({ to: '/403' })
    }
  },
  component: NetOpsLaunch,
})
