export const cashbookKeys = {
  all: ['cashbook'] as const,
  accounts: {
    all: () => [...cashbookKeys.all, 'accounts'] as const,
    list: (activeOnly?: boolean) =>
      [...cashbookKeys.accounts.all(), 'list', { activeOnly: !!activeOnly }] as const,
  },
  categories: {
    all: () => [...cashbookKeys.all, 'categories'] as const,
    list: (activeOnly?: boolean) =>
      [...cashbookKeys.categories.all(), 'list', { activeOnly: !!activeOnly }] as const,
  },
  transactions: {
    all: () => [...cashbookKeys.all, 'transactions'] as const,
    list: (filter?: unknown) =>
      [...cashbookKeys.transactions.all(), 'list', filter || {}] as const,
  },
  balances: (fromUnix?: number, toUnix?: number) =>
    [...cashbookKeys.all, 'balances', { fromUnix, toUnix }] as const,
}
