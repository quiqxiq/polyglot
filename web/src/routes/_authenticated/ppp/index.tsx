import { createFileRoute, redirect } from '@tanstack/react-router'
import { z } from 'zod'
import { PPP } from '@/features/ppp'
import { useAuthStore } from '@/stores/auth-store'
import { canAnyPermission } from '@/hooks/use-can'

const pppSearchSchema = z.object({
  tab: z.enum(['secrets', 'active', 'inactive', 'profiles']).catch('secrets'),
  profile: z.string().optional().catch(''),
  service: z.string().optional().catch(''),
  comment: z.string().optional().catch(''),
  filter: z.string().optional().catch(''),
})

export const Route = createFileRoute('/_authenticated/ppp/')({
  validateSearch: (search) => pppSearchSchema.parse(search),
  beforeLoad: () => {
    const permissions = useAuthStore.getState().auth.user?.permissions
    if (
      !canAnyPermission(permissions, [
        'ppp:read',
        'ppp:secret:read',
        'ppp:active:read',
        'ppp:inactive:read',
        'ppp:profile:read',
      ])
    ) {
      throw redirect({ to: '/403' })
    }
  },
  component: PPP,
})
