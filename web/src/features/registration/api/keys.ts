export const registrationKeys = {
  all: ['registrations'] as const,
  list: (status?: string, phone?: string) =>
    [...registrationKeys.all, 'list', { status: status || 'all', phone: phone || 'all' }] as const,
  detail: (id: string) => [...registrationKeys.all, 'detail', id] as const,
}
