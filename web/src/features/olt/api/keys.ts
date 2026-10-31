export const oltKeys = {
  all: ['olt'] as const,
  lists: () => [...oltKeys.all, 'list'] as const,
  detail: (id: string) => [...oltKeys.all, 'detail', id] as const,
  onus: (oltDeviceId: string, ponIndex: number, state: string) =>
    [...oltKeys.detail(oltDeviceId), 'onus', ponIndex, state] as const,
  onu: (onuId: string) => [...oltKeys.all, 'onu', onuId] as const,
  onuEvents: (onuId: string) => [...oltKeys.onu(onuId), 'events'] as const,
  onuOpticalHistory: (onuId: string, hours: number) =>
    [...oltKeys.onu(onuId), 'optical', hours] as const,
}
