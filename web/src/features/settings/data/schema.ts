import { z } from 'zod'

export const botSettingsSchema = z.object({
  burstLimit: z.number().min(1, 'Burst limit minimal 1'),
  burstWindowSecs: z.number().min(1, 'Burst window minimal 1 detik'),
  mute1hSecs: z.number().min(60, 'Durasi mute minimal 60 detik'),
  ban24hSecs: z.number().min(3600, 'Durasi ban minimal 3600 detik'),
  dailyChatLimit: z.number().min(1, 'Daily chat limit minimal 1'),
  sessionTimeoutMinutes: z.number().min(1, 'Session timeout minimal 1 menit'),
  slidingWindowSize: z.number().min(1, 'Sliding window minimal 1'),
  llmMaxOutputTokens: z.number().min(64, 'Max output tokens minimal 64'),
  whitelistAllStaff: z.boolean().default(true),
  customWhitelistPhones: z.string().default(''),
})

export type BotSettingsFormValues = z.infer<typeof botSettingsSchema>

export const settingItemSchema = z.object({
  key: z.string().min(1, 'Key setting wajib ada'),
  value: z.string().default(''),
  category: z.string().default('general'),
  description: z.string().default(''),
})

export type SettingItemFormValues = z.infer<typeof settingItemSchema>
