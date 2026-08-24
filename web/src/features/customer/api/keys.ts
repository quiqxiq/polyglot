export const customerKeys = {
  all: ['customers'] as const,
  list: () => [...customerKeys.all, 'list'] as const,
  detail: (id: string) => [...customerKeys.all, 'detail', id] as const,
  lookup: (type: 'phone' | 'code' | 'portal', value: string) =>
    [...customerKeys.all, 'lookup', type, value] as const,
  reconcile: (deviceId: string) => [...customerKeys.all, 'reconcile', deviceId] as const,
  routerPreview: (deviceId: string) => [...customerKeys.all, 'routerPreview', deviceId] as const,
}
