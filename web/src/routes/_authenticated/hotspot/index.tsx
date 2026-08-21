import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { Hotspot } from '@/features/hotspot'

const hotspotSearchSchema = z.object({
  tab: z.enum(['users', 'profiles', 'active', 'inactive', 'hosts', 'bindings', 'cookies']).catch('users'),
  profile: z.string().optional().catch(''),
  comment: z.string().optional().catch(''),
  filter: z.string().optional().catch(''),
})

export const Route = createFileRoute('/_authenticated/hotspot/')({
  validateSearch: (search) => hotspotSearchSchema.parse(search),
  component: Hotspot,
})
