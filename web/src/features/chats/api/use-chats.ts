import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type {
  TakeOverConversationRequest,
  ResetConversationBotRequest,
  CloseConversationRequest,
  ResetRateLimitRequest,
} from '@/gen/v1/bot_pb'
import type {
  CreateWASessionRequest,
  MarkChatReadRequest,
  SendWATextMessageRequest,
  ToggleChatBotRequest,
  ToggleWABotRequest,
} from '@/gen/v1/whatsapp_pb'
import { botClient, whatsappClient } from '@/lib/api-client'
import { botKeys } from './keys'

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

export function useConversationContextQuery(id: string, enabled = true) {
  return useQuery({
    queryKey: botKeys.conversationContext(id),
    queryFn: async () => {
      const res = await botClient.getConversationContext({ id })
      return res
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
      queryClient.invalidateQueries({
        queryKey: botKeys.conversation(variables.id),
      })
      queryClient.invalidateQueries({ queryKey: botKeys.conversations() })
      queryClient.invalidateQueries({
        queryKey: botKeys.conversationContext(variables.id),
      })
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
      queryClient.invalidateQueries({
        queryKey: botKeys.conversation(variables.id),
      })
      queryClient.invalidateQueries({ queryKey: botKeys.conversations() })
      queryClient.invalidateQueries({
        queryKey: botKeys.conversationContext(variables.id),
      })
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
      queryClient.invalidateQueries({
        queryKey: botKeys.conversation(variables.id),
      })
      queryClient.invalidateQueries({ queryKey: botKeys.conversations() })
      queryClient.invalidateQueries({
        queryKey: botKeys.conversationContext(variables.id),
      })
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
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: SendWATextMessageRequest) => {
      return await whatsappClient.sendTextMessage(req)
    },
    onSuccess: (_, variables) => {
      // Refresh both the chat list (preview/order) and the conversation pane.
      queryClient.invalidateQueries({
        queryKey: botKeys.chats(variables.sessionId),
      })
      if (variables.recipientPhone) {
        queryClient.invalidateQueries({
          queryKey: botKeys.chatMessages(
            variables.sessionId,
            variables.recipientPhone
          ),
        })
      }
    },
  })
}

// ─── Inbox (chat mirror) ───────────────────────────────────────────────

export function useListChatsQuery(sessionId: string, search = '') {
  return useQuery({
    queryKey: botKeys.chats(sessionId, search),
    queryFn: async () => {
      // Limit dinaikkan (200) supaya filter tab (unread/read/group/new) yang
      // diterapkan client-side tidak kehilangan chat di luar 50 terbaru.
      const res = await whatsappClient.listChats({
        sessionId,
        search,
        limit: 200,
        offset: 0,
      })
      return res.chats
    },
    enabled: Boolean(sessionId),
  })
}

export function useGetChatMessagesQuery(
  sessionId: string,
  chatJid: string,
  enabled = true
) {
  return useQuery({
    queryKey: botKeys.chatMessages(sessionId, chatJid),
    queryFn: async () => {
      const res = await whatsappClient.getChatMessages({
        sessionId,
        chatJid,
        limit: 200,
        offset: 0,
      })
      return res.messages
    },
    enabled: Boolean(sessionId) && Boolean(chatJid) && enabled,
  })
}

export function useMarkChatReadMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: MarkChatReadRequest) => {
      return await whatsappClient.markChatRead(req)
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: botKeys.chats(variables.sessionId),
      })
    },
  })
}

export function useToggleChatBotMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: ToggleChatBotRequest) => {
      return await whatsappClient.toggleChatBot(req)
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: botKeys.chats(variables.sessionId),
      })
    },
  })
}

export function useRateLimitStatusQuery(phoneNumber: string, enabled = true) {
  return useQuery({
    queryKey: botKeys.rateLimitStatus(phoneNumber),
    queryFn: async () => {
      return await botClient.getRateLimitStatus({ phoneNumber })
    },
    enabled: Boolean(phoneNumber) && enabled,
    refetchInterval: 10000,
  })
}

export function useResetRateLimitMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: ResetRateLimitRequest) => {
      return await botClient.resetRateLimit(req)
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: botKeys.rateLimitStatus(variables.phoneNumber),
      })
    },
  })
}

