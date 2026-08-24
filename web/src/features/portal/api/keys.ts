export const portalKeys = {
  all: ['portal'] as const,
  overview: (token: string) => [...portalKeys.all, 'overview', token] as const,
  invoices: (token: string, limit?: number) => [...portalKeys.all, 'invoices', token, limit || 50] as const,
  payments: (token: string, limit?: number) => [...portalKeys.all, 'payments', token, limit || 50] as const,
}
