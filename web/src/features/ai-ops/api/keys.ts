export const aiOpsKeys = {
  all: ['ai-ops'] as const,
  approvals: {
    all: () => [...aiOpsKeys.all, 'approvals'] as const,
    list: (filter: Record<string, unknown>) =>
      [...aiOpsKeys.approvals.all(), 'list', filter] as const,
    detail: (id: string) =>
      [...aiOpsKeys.approvals.all(), 'detail', id] as const,
  },
  restorePoints: {
    all: () => [...aiOpsKeys.all, 'restore-points'] as const,
    list: (filter: Record<string, unknown>) =>
      [...aiOpsKeys.restorePoints.all(), 'list', filter] as const,
    detail: (id: string) =>
      [...aiOpsKeys.restorePoints.all(), 'detail', id] as const,
  },
}
