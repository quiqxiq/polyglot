import { z } from 'zod'

export const requestOtpSchema = z.object({
  identifier: z.string().min(1, 'Masukkan nomor WhatsApp atau kode portal Anda'),
})

export type RequestOtpFormValues = z.infer<typeof requestOtpSchema>

export const portalLoginSchema = z.object({
  identifier: z.string().min(1, 'Masukkan nomor WhatsApp atau kode portal'),
  otp: z.string().min(4, 'Kode OTP minimal 4 digit').max(6, 'Kode OTP maksimal 6 digit'),
})

export type PortalLoginFormValues = z.infer<typeof portalLoginSchema>
