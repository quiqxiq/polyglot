export const userKeys = {
  all: ['users'] as const,
  list: (search?: string) =>
    [...userKeys.all, 'list', search || 'all'] as const,
}
