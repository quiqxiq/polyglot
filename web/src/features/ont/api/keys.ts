export const ontKeys = {
  all: ['ont'] as const,
  acsConfig: () => [...ontKeys.all, 'acs-config'] as const,
  acsDevices: (onlyUnmatched: boolean, search: string) =>
    [...ontKeys.all, 'acs-devices', onlyUnmatched, search] as const,
  customerOnus: (customerId: string) =>
    [...ontKeys.all, 'customer-onus', customerId] as const,
  unlinkedOnus: (oltDeviceId: string) =>
    [...ontKeys.all, 'unlinked-onus', oltDeviceId] as const,
}
