export const notificationKeys = {
  all: ['notifications'] as const,
  templates: {
    all: () => [...notificationKeys.all, 'templates'] as const,
    list: (activeOnly?: boolean) =>
      [...notificationKeys.templates.all(), 'list', { activeOnly: !!activeOnly }] as const,
    detail: (templateKey: string) =>
      [...notificationKeys.templates.all(), 'detail', templateKey] as const,
  },
  queue: {
    all: () => [...notificationKeys.all, 'queue'] as const,
    list: (customerId?: string, status?: string, limit?: number) =>
      [...notificationKeys.queue.all(), 'list', { customerId: customerId || 'all', status: status || 'all', limit: limit || 50 }] as const,
    pendingCount: () => [...notificationKeys.queue.all(), 'pendingCount'] as const,
  },
}
