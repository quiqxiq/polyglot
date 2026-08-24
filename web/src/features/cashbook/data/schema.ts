import { z } from 'zod'

export const cashAccountSchema = z.object({
  id: z.string().optional(),
  accountCode: z.string().min(1, 'Kode akun wajib diisi'),
  name: z.string().min(1, 'Nama rekening wajib diisi'),
  type: z.enum(['CASH', 'BANK']).default('CASH'),
  isActive: z.boolean().default(true),
})

export type CashAccountFormValues = z.infer<typeof cashAccountSchema>

export const cashCategorySchema = z.object({
  id: z.string().optional(),
  name: z.string().min(1, 'Nama kategori wajib diisi'),
  type: z.enum(['INCOME', 'EXPENSE']).default('INCOME'),
  isActive: z.boolean().default(true),
})

export type CashCategoryFormValues = z.infer<typeof cashCategorySchema>

export const addTransactionSchema = z.object({
  accountId: z.string().min(1, 'Pilih rekening kas'),
  categoryId: z.string().min(1, 'Pilih kategori transaksi'),
  direction: z.enum(['IN', 'OUT']),
  amount: z.number().positive('Nominal transaksi harus lebih dari 0'),
  description: z.string().min(1, 'Deskripsi transaksi wajib diisi'),
})

export type AddTransactionFormValues = z.infer<typeof addTransactionSchema>
