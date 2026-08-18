import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { PPP } from '@/features/ppp'

const pppSearchSchema = z.object({
  tab: z.enum(['secrets', 'active', 'inactive', 'profiles']).catch('secrets'),
  profile: z.string().optional().catch(''),
  service: z.string().optional().catch(''),
  comment: z.string().optional().catch(''),
  filter: z.string().optional().catch(''),
})

export const Route = createFileRoute('/_authenticated/ppp/')({
  validateSearch: (search) => pppSearchSchema.parse(search),
  component: PPP,
})
