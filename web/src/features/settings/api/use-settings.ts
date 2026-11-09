import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { settingClient } from '@/lib/api-client'
import {
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

// Backwards compatibility alias
export const useGetAllSettings = useGetAllSettingsQuery
