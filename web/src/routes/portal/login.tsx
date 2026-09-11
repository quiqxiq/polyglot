import { createFileRoute } from '@tanstack/react-router'
import { PortalLoginPage } from '@/features/portal/components/portal-login-page'

export const Route = createFileRoute('/portal/login')({
  component: PortalLoginPage,
})
