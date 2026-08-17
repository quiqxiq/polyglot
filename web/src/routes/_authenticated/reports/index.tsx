import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { Reports } from '@/features/reports'

const reportsSearchSchema = z.object({
  day: z.string().optional().catch(''),
  month: z.string().optional().catch(''),
  year: z.string().optional().catch(''),
  filter: z.string().optional().catch(''),
})

export const Route = createFileRoute('/_authenticated/reports/')({
  validateSearch: (search) => reportsSearchSchema.parse(search),
  component: Reports,
})
