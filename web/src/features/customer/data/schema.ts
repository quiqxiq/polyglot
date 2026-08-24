import { z } from 'zod'

export const customerFormSchema = z.object({
  id: z.string().optional(),
  name: z.string().min(1, 'Nama pelanggan wajib diisi'),
  phone: z.string().min(8, 'Nomor HP minimal 8 digit'),
  email: z.string().email('Format email tidak valid').or(z.literal('')),
  address: z.string().default(''),
  latitude: z.number().default(0),
  longitude: z.number().default(0),
  hasCoordinates: z.boolean().default(false),
  status: z.enum(['ACTIVE', 'ISOLATED', 'SUSPENDED', 'TERMINATED']).default('ACTIVE'),
  notes: z.string().default(''),
})

export type CustomerFormValues = z.infer<typeof customerFormSchema>

export const importRouterSchema = z.object({
  deviceId: z.string().min(1, 'Pilih router target'),
  deviceName: z.string().default(''),
  dryRun: z.boolean().default(true),
})

export type ImportRouterFormValues = z.infer<typeof importRouterSchema>

export const exportCustomersSchema = z.object({
  format: z.number().default(0),
})

export type ExportCustomersFormValues = z.infer<typeof exportCustomersSchema>
