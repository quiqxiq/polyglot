import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { botClient, whatsappClient } from '@/lib/api-client'
import { botKeys } from './keys'
import {
  TakeOverConversationRequest,
  ResetConversationBotRequest,
  CloseConversationRequest,
} from '@/gen/v1/bot_pb'
import {
  CreateWASessionRequest,
  SendWATextMessageRequest,
  ToggleWABotRequest,
} from '@/gen/v1/whatsapp_pb'

export function useConversationsQuery(sessionId = '') {
  return useQuery({
    queryKey: botKeys.conversations(sessionId),
    queryFn: async () => {
      const res = await botClient.listConversations({ sessionId })
      return res.conversations
    },
  })
}

export function useConversationQuery(id: string, enabled = true) {
  return useQuery({
    queryKey: botKeys.conversation(id),
    queryFn: async () => {
      const res = await botClient.getConversation({ id })
      return res.conversation
    },
    enabled: Boolean(id) && enabled,
  })
}

export function useTakeOverConversationMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: TakeOverConversationRequest) => {
      return await botClient.takeOverConversation(req)
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: botKeys.conversation(variables.id) })
      queryClient.invalidateQueries({ queryKey: botKeys.conversations() })
    },
  })
}

export function useResetConversationBotMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: ResetConversationBotRequest) => {
      return await botClient.resetConversationBot(req)
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: botKeys.conversation(variables.id) })
      queryClient.invalidateQueries({ queryKey: botKeys.conversations() })
    },
  })
}

export function useCloseConversationMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: CloseConversationRequest) => {
      return await botClient.closeConversation(req)
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: botKeys.conversation(variables.id) })
      queryClient.invalidateQueries({ queryKey: botKeys.conversations() })
    },
  })
}

// WhatsApp Sessions
export function useWASessionsQuery() {
  return useQuery({
    queryKey: botKeys.waSessions(),
    queryFn: async () => {
      const res = await whatsappClient.listSessions({})
      return res.sessions
    },
  })
}

export function useCreateWASessionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: CreateWASessionRequest) => {
      return await whatsappClient.createSession(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: botKeys.waSessions() })
    },
  })
}

export function useWASessionQRQuery(sessionId: string, enabled = true) {
  return useQuery({
    queryKey: botKeys.waQR(sessionId),
    queryFn: async () => {
      const res = await whatsappClient.getQRCode({ sessionId })
      return res
    },
    enabled: Boolean(sessionId) && enabled,
  })
}

export function useToggleWABotMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: ToggleWABotRequest) => {
      return await whatsappClient.toggleBot(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: botKeys.waSessions() })
    },
  })
}

export function useSendWATextMessageMutation() {
  return useMutation({
    mutationFn: async (req: SendWATextMessageRequest) => {
      return await whatsappClient.sendTextMessage(req)
    },
  })
}
