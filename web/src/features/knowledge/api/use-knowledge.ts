import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { knowledgeClient } from '@/lib/api-client'
import { knowledgeKeys } from './keys'
import {
  CreateKnowledgeRequest,
  UpdateKnowledgeRequest,
  DeleteKnowledgeRequest,
  CreateLLMConfigRequest,
  UpdateLLMConfigRequest,
  ActivateLLMConfigRequest,
  TestLLMConfigRequest,
  DeleteLLMConfigRequest,
  CreateTechnicianRequest,
  UpdateTechnicianRequest,
  ToggleTechnicianActiveRequest,
  DeleteTechnicianRequest,
} from '@/gen/v1/knowledge_pb'

// Knowledge Items
export function useKnowledgeListQuery(category = '', searchQuery = '') {
  return useQuery({
    queryKey: knowledgeKeys.items(category, searchQuery),
    queryFn: async () => {
      const res = await knowledgeClient.listKnowledge({ category, searchQuery })
      return res.items
    },
  })
}

export function useCreateKnowledgeMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: CreateKnowledgeRequest) => {
      return await knowledgeClient.createKnowledge(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: knowledgeKeys.items() })
    },
  })
}

export function useUpdateKnowledgeMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: UpdateKnowledgeRequest) => {
      return await knowledgeClient.updateKnowledge(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: knowledgeKeys.items() })
    },
  })
}

export function useDeleteKnowledgeMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: DeleteKnowledgeRequest) => {
      return await knowledgeClient.deleteKnowledge(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: knowledgeKeys.items() })
    },
  })
}

// LLM Configs
export function useLLMConfigsQuery() {
  return useQuery({
    queryKey: knowledgeKeys.llmConfigs(),
    queryFn: async () => {
      const res = await knowledgeClient.listLLMConfigs({})
      return res.configs
    },
  })
}

export function useCreateLLMConfigMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: CreateLLMConfigRequest) => {
      return await knowledgeClient.createLLMConfig(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: knowledgeKeys.llmConfigs() })
    },
  })
}

export function useUpdateLLMConfigMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: UpdateLLMConfigRequest) => {
      return await knowledgeClient.updateLLMConfig(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: knowledgeKeys.llmConfigs() })
    },
  })
}

export function useActivateLLMConfigMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: ActivateLLMConfigRequest) => {
      return await knowledgeClient.activateLLMConfig(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: knowledgeKeys.llmConfigs() })
    },
  })
}

export function useTestLLMConfigMutation() {
  return useMutation({
    mutationFn: async (req: TestLLMConfigRequest) => {
      return await knowledgeClient.testLLMConfig(req)
    },
  })
}

export function useDeleteLLMConfigMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: DeleteLLMConfigRequest) => {
      return await knowledgeClient.deleteLLMConfig(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: knowledgeKeys.llmConfigs() })
    },
  })
}

// Technicians
export function useTechniciansQuery() {
  return useQuery({
    queryKey: knowledgeKeys.technicians(),
    queryFn: async () => {
      const res = await knowledgeClient.listTechnicians({})
      return res.technicians
    },
  })
}

export function useCreateTechnicianMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: CreateTechnicianRequest) => {
      return await knowledgeClient.createTechnician(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: knowledgeKeys.technicians() })
    },
  })
}

export function useUpdateTechnicianMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: UpdateTechnicianRequest) => {
      return await knowledgeClient.updateTechnician(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: knowledgeKeys.technicians() })
    },
  })
}

export function useToggleTechnicianActiveMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: ToggleTechnicianActiveRequest) => {
      return await knowledgeClient.toggleTechnicianActive(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: knowledgeKeys.technicians() })
    },
  })
}

export function useDeleteTechnicianMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: DeleteTechnicianRequest) => {
      return await knowledgeClient.deleteTechnician(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: knowledgeKeys.technicians() })
    },
  })
}
