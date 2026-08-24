export const settingsKeys = {
  all: ['settings'] as const,
  list: (category?: string) => [...settingsKeys.all, 'list', category || 'all'] as const,
  bot: () => [...settingsKeys.all, 'bot'] as const,
}
