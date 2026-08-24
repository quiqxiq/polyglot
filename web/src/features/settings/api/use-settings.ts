import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { settingClient } from '@/lib/api-client'
import {
  type BotSettings,
  type UpdateSettingRequest,
  type BatchUpdateSettingsRequest,
} from '@/gen/v1/settings_pb'
import { settingsKeys } from './keys'

export function useGetAllSettingsQuery() {
  return useQuery({
    queryKey: settingsKeys.all,
    queryFn: async () => {
      const res = await settingClient.getAllSettings({})
      return res.settings
    },
  })
}

export function useSettingsByCategoryQuery(category: string) {
  return useQuery({
    queryKey: settingsKeys.list(category),
    queryFn: async () => {
      const res = await settingClient.getSettingsByCategory({ category })
      return res.settings
    },
    enabled: Boolean(category),
  })
}

export function useBotSettingsQuery() {
  return useQuery({
    queryKey: settingsKeys.bot(),
    queryFn: async () => {
      const res = await settingClient.getBotSettings({})
      return res.settings
    },
  })
}

export function useUpdateSettingMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: UpdateSettingRequest) => {
      return await settingClient.updateSetting(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: settingsKeys.all })
      toast.success('Pengaturan berhasil diperbarui!')
    },
    onError: (err: Error) => {
      toast.error(`Gagal memperbarui pengaturan: ${err.message}`)
    },
  })
}

export function useBatchUpdateSettingsMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: BatchUpdateSettingsRequest) => {
      return await settingClient.batchUpdateSettings(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: settingsKeys.all })
      toast.success('Pengaturan batch berhasil disimpan!')
    },
    onError: (err: Error) => {
      toast.error(`Gagal menyimpan pengaturan: ${err.message}`)
    },
  })
}

export function useUpdateBotSettingsMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (settings: Partial<BotSettings>) => {
      const res = await settingClient.updateBotSettings({
        settings: settings as BotSettings,
      })
      return res.settings
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: settingsKeys.bot() })
      toast.success('Pengaturan bot & anti-spam berhasil disimpan!')
    },
    onError: (err: Error) => {
      toast.error(`Gagal menyimpan pengaturan bot: ${err.message}`)
    },
  })
}

// Backwards compatibility aliases
export const useGetBotSettings = useBotSettingsQuery
export const useUpdateBotSettings = useUpdateBotSettingsMutation
export const useGetAllSettings = useGetAllSettingsQuery
