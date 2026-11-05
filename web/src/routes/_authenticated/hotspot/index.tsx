import { createFileRoute, redirect } from '@tanstack/react-router'
import { z } from 'zod'
import { Hotspot } from '@/features/hotspot'
import { useAuthStore } from '@/stores/auth-store'
import { canAnyPermission } from '@/hooks/use-can'

const hotspotSearchSchema = z.object({
  tab: z.enum(['users', 'profiles', 'active', 'inactive', 'hosts', 'bindings', 'cookies']).catch('users'),
  profile: z.string().optional().catch(''),
  comment: z.string().optional().catch(''),
  filter: z.string().optional().catch(''),
})

export const Route = createFileRoute('/_authenticated/hotspot/')({
  validateSearch: (search) => hotspotSearchSchema.parse(search),
  beforeLoad: () => {
    const permissions = useAuthStore.getState().auth.user?.permissions
    if (
      !canAnyPermission(permissions, [
        'hotspot:read',
        'hotspot:user:read',
        'hotspot:profile:read',
        'hotspot:active:read',
        'hotspot:host:read',
        'hotspot:binding:read',
      ])
    ) {
      throw redirect({ to: '/403' })
    }
  },
  component: Hotspot,
})
