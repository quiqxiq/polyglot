import { createFileRoute, redirect } from '@tanstack/react-router'
import { WhatsAppDevices } from '@/features/whatsapp'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'

export const Route = createFileRoute('/_authenticated/whatsapp/')({
  beforeLoad: () => {
    const permissions = useAuthStore.getState().auth.user?.permissions
    if (!canPermission(permissions, 'whatsapp:read')) {
      throw redirect({ to: '/403' })
    }
  },
  component: WhatsAppDevices,
})
