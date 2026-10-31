import { createFileRoute, redirect } from '@tanstack/react-router'
import { ApprovalInbox } from '@/features/ai-ops/components/approval-inbox'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'

export const Route = createFileRoute('/_authenticated/ai-ops/approvals')({
  beforeLoad: () => {
    const permissions = useAuthStore.getState().auth.user?.permissions
    if (!canPermission(permissions, 'approval:read')) {
      throw redirect({ to: '/403' })
    }
  },
  component: ApprovalInbox,
})
