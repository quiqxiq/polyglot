import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { Dashboard } from '@/features/dashboard'

const dashboardSearchSchema = z.object({
  tab: z.enum(['utama', 'laporan', 'trend', 'kpi', 'lainnya']).catch('utama').optional(),
})

export const Route = createFileRoute('/_authenticated/')({
  validateSearch: (search) => dashboardSearchSchema.parse(search),
  component: Dashboard,
})

