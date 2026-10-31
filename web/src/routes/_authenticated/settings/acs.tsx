import { createFileRoute, redirect } from '@tanstack/react-router'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'
import { SettingsAcs } from '@/features/settings/acs'

export const Route = createFileRoute('/_authenticated/settings/acs')({
  beforeLoad: () => {
    const permissions = useAuthStore.getState().auth.user?.permissions
    if (!canPermission(permissions, 'ont:read')) {
      throw redirect({ to: '/403' })
    }
  },
  component: SettingsAcs,
})
