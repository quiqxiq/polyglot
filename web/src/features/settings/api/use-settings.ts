import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { settingClient } from '@/lib/api-client'
import { BotSettings } from '@/gen/v1/settings_pb'
import { toast } from 'sonner'

export function useGetBotSettings() {
  return useQuery({
    queryKey: ['settings', 'bot'],
    queryFn: async () => {
      const res = await settingClient.getBotSettings({})
      return res.settings
    },
  })
}

export function useUpdateBotSettings() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (settings: Partial<BotSettings>) => {
      const payload = new BotSettings(settings as any)
      const res = await settingClient.updateBotSettings({
        settings: payload,
      })
      return res.settings
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings', 'bot'] })
      toast.success('Pengaturan bot & anti-spam berhasil disimpan!')
    },
    onError: (err: Error) => {
      toast.error(`Gagal menyimpan pengaturan bot: ${err.message}`)
    },
  })
}

export function useGetAllSettings() {
  return useQuery({
    queryKey: ['settings', 'all'],
    queryFn: async () => {
      const res = await settingClient.getAllSettings({})
      return res.settings
    },
  })
}
