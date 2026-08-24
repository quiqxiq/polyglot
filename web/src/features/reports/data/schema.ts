import { z } from 'zod'

export const reportFilterSchema = z.object({
  periodType: z.enum(['DAILY', 'MONTHLY', 'YEARLY']).default('MONTHLY'),
  date: z.string().default(''),
  month: z.string().default(''),
  year: z.number().default(new Date().getFullYear()),
})

export type ReportFilterFormValues = z.infer<typeof reportFilterSchema>

export const refreshSnapshotSchema = z.object({
  date: z.string().min(1, 'Pilih tanggal snapshot yang ingin direbuild'),
})

export type RefreshSnapshotFormValues = z.infer<typeof refreshSnapshotSchema>
