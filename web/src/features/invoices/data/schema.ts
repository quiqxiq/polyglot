import { z } from 'zod'

export const cashierPaySchema = z.object({
  invoiceId: z.string().min(1, 'Invoice wajib dipilih'),
  amount: z.number().positive('Nominal pembayaran harus lebih dari 0'),
  cashAccountId: z.string().min(1, 'Pilih rekening kas penampung'),
  incomeCategoryId: z.string().min(1, 'Pilih kategori pos pendapatan'),
  scanMethod: z.string(),
  reference: z.string(),
  notes: z.string(),
})

export type CashierPayFormValues = z.infer<typeof cashierPaySchema>

export const generateInvoicesSchema = z.object({
  period: z.string(),
})

export type GenerateInvoicesFormValues = z.infer<typeof generateInvoicesSchema>
