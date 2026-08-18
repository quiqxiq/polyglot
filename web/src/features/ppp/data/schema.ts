import { z } from 'zod'

export const pppSecretSchema = z.object({
  name: z.string().min(1, 'Username is required'),
  password: z.string(),
  profile: z.string().min(1, 'Profile is required'),
  service: z.string(),
  localAddress: z.string(),
  remoteAddress: z.string(),
  comment: z.string(),
  callerId: z.string(),
  disabled: z.boolean(),
})

export type PPPSecretFormValues = z.infer<typeof pppSecretSchema>

export const pppProfileSchema = z.object({
  name: z.string().min(1, 'Profile name is required'),
  rateLimit: z.string(),
  localAddress: z.string(),
  remoteAddress: z.string(),
  dnsServer: z.string(),
  parentQueue: z.string(),
  addressList: z.string(),
  comment: z.string(),
  sharedUsers: z.string(),
  onlyOne: z.enum(['default', 'yes', 'no']),
})

export type PPPProfileFormValues = z.infer<typeof pppProfileSchema>
