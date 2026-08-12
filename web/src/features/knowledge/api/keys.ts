export const knowledgeKeys = {
  all: ['knowledge'] as const,
  items: (category?: string, query?: string) => [...knowledgeKeys.all, 'items', category || 'all', query || 'all'] as const,
  llmConfigs: () => [...knowledgeKeys.all, 'llm-configs'] as const,
  technicians: () => [...knowledgeKeys.all, 'technicians'] as const,
}
