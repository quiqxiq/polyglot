export const botKeys = {
  all: ['bot'] as const,
  conversations: (sessionId?: string) => [...botKeys.all, 'conversations', sessionId || 'all'] as const,
  conversation: (id: string) => [...botKeys.all, 'conversation', id] as const,
  conversationContext: (id: string) => [...botKeys.all, 'conversation-context', id] as const,
  waSessions: () => [...botKeys.all, 'wa-sessions'] as const,
  waQR: (sessionId: string) => [...botKeys.all, 'wa-qr', sessionId] as const,
  chats: (sessionId: string, search = '') => [...botKeys.all, 'chats', sessionId, search] as const,
  chatMessages: (sessionId: string, chatJid: string) => [...botKeys.all, 'chat-messages', sessionId, chatJid] as const,
  rateLimitStatus: (phoneNumber: string) => [...botKeys.all, 'rate-limit-status', phoneNumber] as const,
}

