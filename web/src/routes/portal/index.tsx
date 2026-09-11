import { createFileRoute, redirect } from '@tanstack/react-router'
import { PortalDashboard } from '@/features/portal/components/portal-dashboard'
import { getPortalToken } from '@/features/portal/api/portal-http'

export const Route = createFileRoute('/portal/')({
  beforeLoad: () => {
    if (!getPortalToken()) {
      throw redirect({ to: '/portal/login' })
    }
  },
  component: PortalDashboard,
})
