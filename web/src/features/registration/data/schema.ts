import { z } from 'zod'

export const submitRegistrationSchema = z.object({
  fullName: z.string().min(1, 'Nama lengkap wajib diisi'),
  phone: z.string().min(8, 'Nomor WhatsApp minimal 8 digit'),
  email: z.string().email('Format email tidak valid').or(z.literal('')),
  address: z.string().min(1, 'Alamat pemasangan wajib diisi'),
  latitude: z.number().default(0),
  longitude: z.number().default(0),
  hasCoordinates: z.boolean().default(false),
  notes: z.string().default(''),
  planId: z.string().min(1, 'Pilih paket internet yang diinginkan'),
})

export type SubmitRegistrationFormValues = z.infer<typeof submitRegistrationSchema>

export const approveRegistrationSchema = z.object({
  id: z.string().min(1, 'ID registrasi wajib ada'),
  adminNotes: z.string().default(''),
})

export type ApproveRegistrationFormValues = z.infer<typeof approveRegistrationSchema>

export const scheduleInstallSchema = z.object({
  id: z.string().min(1, 'ID registrasi wajib ada'),
  installDateUnix: z.number().positive('Pilih tanggal pemasangan'),
  installTimeHhmm: z.string().default('09:00'),
  technicianId: z.string().default(''),
})

export type ScheduleInstallFormValues = z.infer<typeof scheduleInstallSchema>

export const markInstalledSchema = z.object({
  id: z.string().min(1, 'ID registrasi wajib ada'),
  technicianNotes: z.string().min(1, 'Catatan teknisi / redaman ODP wajib diisi'),
})

export type MarkInstalledFormValues = z.infer<typeof markInstalledSchema>

export const convertRegistrationSchema = z.object({
  id: z.string().min(1, 'ID registrasi wajib ada'),
  deviceId: z.string().default(''),
})

export type ConvertRegistrationFormValues = z.infer<typeof convertRegistrationSchema>

export const rejectRegistrationSchema = z.object({
  id: z.string().min(1, 'ID registrasi wajib ada'),
  reason: z.string().min(1, 'Alasan penolakan wajib diisi'),
})

export type RejectRegistrationFormValues = z.infer<typeof rejectRegistrationSchema>

export const cancelRegistrationSchema = z.object({
  id: z.string().min(1, 'ID registrasi wajib ada'),
  reason: z.string().min(1, 'Alasan pembatalan wajib diisi'),
})

export type CancelRegistrationFormValues = z.infer<typeof cancelRegistrationSchema>
