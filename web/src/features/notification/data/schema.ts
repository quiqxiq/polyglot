import { z } from 'zod'

export const notificationTemplateSchema = z.object({
  id: z.string().optional(),
  templateKey: z.string().min(1, 'Key template wajib diisi (contoh: BILL_REMINDER)'),
  name: z.string().min(1, 'Nama template wajib diisi'),
  content: z.string().min(1, 'Konten pesan template wajib diisi'),
  variablesJson: z.string().default('[]'),
  isActive: z.boolean().default(true),
})

export type NotificationTemplateFormValues = z.infer<typeof notificationTemplateSchema>

export const testSendSchema = z.object({
  phone: z.string().min(8, 'Nomor WhatsApp tujuan minimal 8 digit'),
  content: z.string().min(1, 'Isi pesan uji coba tidak boleh kosong'),
})

export type TestSendFormValues = z.infer<typeof testSendSchema>
