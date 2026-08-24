import { z } from 'zod'

export const planFormSchema = z.object({
  id: z.string().optional(),
  name: z.string().min(1, 'Nama paket wajib diisi'),
  serviceType: z.enum(['PPPOE', 'HOTSPOT', 'DEDICATED']),
  bandwidthDownloadKbps: z.number().min(64, 'Download minimal 64 Kbps'),
  bandwidthUploadKbps: z.number().min(64, 'Upload minimal 64 Kbps'),
  burstDownloadKbps: z.number().default(0),
  burstUploadKbps: z.number().default(0),
  burstThresholdKbps: z.number().default(0),
  burstTimeSeconds: z.number().default(0),
  price: z.number().min(0, 'Harga tidak boleh negatif'),
  sellingPrice: z.number().default(0),
  installationFee: z.number().default(0),
  taxPercent: z.number().default(0),
  validity: z.string().default('30d'),
  validityMode: z.enum(['CALENDAR', 'UPTIME']).default('CALENDAR'),
  simultaneousUse: z.number().default(1),
  ipPoolName: z.string().default(''),
  remoteAddressPool: z.string().default(''),
  parentQueue: z.string().default('none'),
  addressList: z.string().default(''),
  sharedUsers: z.number().default(1),
  expireMode: z.string().default('0'),
  lockUser: z.boolean().default(false),
  lockServer: z.boolean().default(false),
  isActive: z.boolean().default(true),
  description: z.string().default(''),
})

export type PlanFormValues = z.infer<typeof planFormSchema>

export const cashierPaySchema = z.object({
  invoiceId: z.string().min(1, 'Invoice ID wajib dipilih'),
  amount: z.number().positive('Nominal pembayaran harus lebih dari 0'),
  cashAccountId: z.string().min(1, 'Pilih rekening kas penampung'),
  incomeCategoryId: z.string().default(''),
  scanMethod: z.string().default('CODE_INPUT'),
  reference: z.string().default(''),
  notes: z.string().default(''),
})

export type CashierPayFormValues = z.infer<typeof cashierPaySchema>

export const changePlanSchema = z.object({
  subscriptionId: z.string().min(1, 'Subscription ID wajib ada'),
  newPlanId: z.string().min(1, 'Pilih paket baru'),
})

export type ChangePlanFormValues = z.infer<typeof changePlanSchema>

export const suspendSubscriptionSchema = z.object({
  subscriptionId: z.string().min(1, 'Subscription ID wajib ada'),
  reason: z.string().min(1, 'Alasan penangguhan wajib diisi'),
})

export type SuspendSubscriptionFormValues = z.infer<typeof suspendSubscriptionSchema>

export const generateInvoicesSchema = z.object({
  period: z.string().regex(/^(\d{4}-\d{2})?$/, 'Format periode harus YYYY-MM atau kosongkan untuk bulan berjalan'),
})

export type GenerateInvoicesFormValues = z.infer<typeof generateInvoicesSchema>
