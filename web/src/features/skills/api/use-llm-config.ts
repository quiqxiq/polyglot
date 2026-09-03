import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { llmConfigClient } from '@/lib/api-client'
import { toast } from 'sonner'
import {
  CreateLLMConfigRequest,
  UpdateLLMConfigRequest,
  ActivateLLMConfigRequest,
  DeleteLLMConfigRequest,
  TestLLMConfigRequest,
} from '@/gen/v1/llm_pb'

export const llmConfigKeys = {
  all: ['llm-configs'] as const,
  lists: () => [...llmConfigKeys.all, 'list'] as const,
}

export function useLLMConfigsQuery() {
  return useQuery({
    queryKey: llmConfigKeys.lists(),
    queryFn: async () => {
      const res = await llmConfigClient.listLLMConfigs({})
      return res.configs
    },
  })
}

export function useCreateLLMConfigMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (req: CreateLLMConfigRequest) => {
      return await llmConfigClient.createLLMConfig(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: llmConfigKeys.lists() })
      toast.success('Konfigurasi LLM berhasil ditambahkan')
    },
    onError: (err: Error) => {
      toast.error(`Gagal membuat konfigurasi LLM: ${err.message}`)
    },
  })
}

export function useUpdateLLMConfigMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (req: UpdateLLMConfigRequest) => {
      return await llmConfigClient.updateLLMConfig(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: llmConfigKeys.lists() })
      toast.success('Konfigurasi LLM berhasil diperbarui')
    },
    onError: (err: Error) => {
      toast.error(`Gagal memperbarui konfigurasi LLM: ${err.message}`)
    },
  })
}

export function useActivateLLMConfigMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (req: ActivateLLMConfigRequest) => {
      return await llmConfigClient.activateLLMConfig(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: llmConfigKeys.lists() })
      toast.success('Model LLM aktif berhasil diubah (Hot Reload Aktif)')
    },
    onError: (err: Error) => {
      toast.error(`Gagal mengaktifkan LLM: ${err.message}`)
    },
  })
}

export function useTestLLMConfigMutation() {
  return useMutation({
    mutationFn: async (req: TestLLMConfigRequest) => {
      return await llmConfigClient.testLLMConfig(req)
    },
  })
}

export function useDeleteLLMConfigMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (req: DeleteLLMConfigRequest) => {
      return await llmConfigClient.deleteLLMConfig(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: llmConfigKeys.lists() })
      toast.success('Konfigurasi LLM berhasil dihapus')
    },
    onError: (err: Error) => {
      toast.error(`Gagal menghapus LLM: ${err.message}`)
    },
  })
}
