import { createFileRoute, redirect } from '@tanstack/react-router'
import { z } from 'zod'
import { LogsFeature } from '@/features/logs'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'

const logsSearchSchema = z.object({
  tab: z.enum(['all', 'hotspot', 'ppp']).catch('all'),
  q: z.string().optional().catch(''),
  severity: z.enum(['all', 'error', 'warning', 'info']).catch('all'),
})

export const Route = createFileRoute('/_authenticated/logs/')({
  beforeLoad: () => {
    const permissions = useAuthStore.getState().auth.user?.permissions
    if (!canPermission(permissions, 'log:read')) {
      throw redirect({ to: '/403' })
    }
  },
  validateSearch: (search) => logsSearchSchema.parse(search),
  component: LogsFeature,
})
