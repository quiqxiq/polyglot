export const botKeys = {
  all: ['bot'] as const,
  conversations: (sessionId?: string) => [...botKeys.all, 'conversations', sessionId || 'all'] as const,
  conversation: (id: string) => [...botKeys.all, 'conversation', id] as const,
  waSessions: () => [...botKeys.all, 'wa-sessions'] as const,
  waQR: (sessionId: string) => [...botKeys.all, 'wa-qr', sessionId] as const,
}
