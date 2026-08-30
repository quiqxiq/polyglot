import { z } from 'zod'

export const submitRegistrationSchema = z.object({
  fullName: z.string().min(2, 'Nama lengkap minimal 2 karakter'),
  phone: z.string().min(8, 'Nomor WhatsApp / telepon wajib diisi'),
  email: z.string().email('Format email tidak valid').or(z.literal('')),
  address: z.string().min(3, 'Alamat pemasangan wajib diisi'),
  planId: z.string().min(1, 'Paket layanan wajib dipilih'),
  latitude: z.coerce.number().optional(),
  longitude: z.coerce.number().optional(),
  notes: z.string().optional(),
})

export type SubmitRegistrationValues = z.infer<typeof submitRegistrationSchema>

export const scheduleInstallSchema = z.object({
  id: z.string().min(1),
  installDate: z.string().min(1, 'Tanggal pasang wajib diisi'),
  installTimeHhmm: z.string().optional(),
  technicianId: z.string().optional(),
  adminNotes: z.string().optional(),
})

export type ScheduleInstallValues = z.infer<typeof scheduleInstallSchema>

export const markInstalledSchema = z.object({
  id: z.string().min(1),
  deviceId: z.string().min(1, 'Router BRAS target wajib dipilih saat pemasangan'),
  technicianNotes: z.string().min(1, 'Catatan teknisi (redaman ODP, port, nomor tiang) wajib diisi'),
})

export type MarkInstalledValues = z.infer<typeof markInstalledSchema>

export const convertRegistrationSchema = z.object({
  id: z.string().min(1),
  deviceId: z.string().min(1, 'Router BRAS target wajib dipilih untuk aktivasi'),
  technicianNotes: z.string().optional(),
})

export type ConvertRegistrationValues = z.infer<typeof convertRegistrationSchema>

export const rejectRegistrationSchema = z.object({
  id: z.string().min(1),
  reason: z.string().min(3, 'Alasan penolakan minimal 3 karakter'),
})

export type RejectRegistrationValues = z.infer<typeof rejectRegistrationSchema>

export const cancelRegistrationSchema = z.object({
  id: z.string().min(1),
  reason: z.string().min(3, 'Alasan pembatalan minimal 3 karakter'),
})

export type CancelRegistrationValues = z.infer<typeof cancelRegistrationSchema>
