import { z } from 'zod'

export const userFormSchema = z.object({
  name: z.string().min(1, 'Name / Username is required'),
  password: z.string().min(1, 'Password is required'),
  profile: z.string().min(1, 'Profile is required'),
  server: z.string(),
  macAddress: z.string(),
  timeLimit: z.string(),
  dataLimit: z.string(),
  comment: z.string(),
  resetCounter: z.boolean(),
})

export type UserFormValues = z.infer<typeof userFormSchema>

export const profileFormSchema = z.object({
  name: z.string().min(1, 'Profile name is required'),
  addressPool: z.string(),
  sharedUsers: z.string(),
  rateLimit: z.string(),
  parentQueue: z.string(),
  price: z.string(),
  sellingPrice: z.string(),
  validity: z.string(),
  expireMode: z.enum(['0', 'rem', 'ntf', 'remc', 'ntfc']),
  lockUser: z.boolean(),
  lockServer: z.boolean(),
  enableRecording: z.boolean(),
  comment: z.string(),
})

export type ProfileFormValues = z.infer<typeof profileFormSchema>

export const voucherGenerateSchema = z.object({
  count: z.number().min(1, 'Quantity must be at least 1').max(500, 'Max 500 vouchers per batch'),
  server: z.string(),
  userType: z.enum(['vc', 'up']),
  userLength: z.number().min(3).max(12),
  prefix: z.string(),
  characterSet: z.string(),
  profile: z.string().min(1, 'Profile is required'),
  timeLimit: z.string(),
  dataLimit: z.string(),
  comment: z.string(),
})

export type VoucherGenerateValues = z.infer<typeof voucherGenerateSchema>

export const expireSetupSchema = z.object({
  interval: z.string().min(1, 'Interval is required'),
})

export type ExpireSetupValues = z.infer<typeof expireSetupSchema>

export const bindingFormSchema = z.object({
  macAddress: z.string(),
  address: z.string(),
  toAddress: z.string(),
  server: z.string(),
  type: z.enum(['bypassed', 'blocked', 'regular']),
  comment: z.string(),
  disabled: z.boolean(),
})

export type BindingFormValues = z.infer<typeof bindingFormSchema>
