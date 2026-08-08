export const billingKeys = {
  all: ['billing'] as const,
  invoices: (customerId?: string) => [...billingKeys.all, 'invoices', customerId || 'all'] as const,
  subscriptions: (customerId?: string) => [...billingKeys.all, 'subscriptions', customerId || 'all'] as const,
}
