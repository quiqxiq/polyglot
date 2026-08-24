export const billingKeys = {
  all: ['billing'] as const,
  invoices: {
    all: () => [...billingKeys.all, 'invoices'] as const,
    list: (customerId?: string, status?: string) =>
      [...billingKeys.invoices.all(), 'list', { customerId: customerId || 'all', status: status || 'all' }] as const,
    detail: (id: string) => [...billingKeys.invoices.all(), 'detail', id] as const,
  },
  subscriptions: {
    all: () => [...billingKeys.all, 'subscriptions'] as const,
    list: (customerId?: string) =>
      [...billingKeys.subscriptions.all(), 'list', customerId || 'all'] as const,
    detail: (id: string) => [...billingKeys.subscriptions.all(), 'detail', id] as const,
  },
  plans: {
    all: () => [...billingKeys.all, 'plans'] as const,
    list: (activeOnly?: boolean) =>
      [...billingKeys.plans.all(), 'list', { activeOnly: !!activeOnly }] as const,
    detail: (id: string) => [...billingKeys.plans.all(), 'detail', id] as const,
  },
  cashier: {
    all: () => [...billingKeys.all, 'cashier'] as const,
    resolve: (identifier: string, method: number) =>
      [...billingKeys.cashier.all(), 'resolve', { identifier, method }] as const,
  },
}
