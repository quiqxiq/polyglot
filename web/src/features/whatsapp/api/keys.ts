export const waDeviceKeys = {
  all: ['wa-devices'] as const,
  sessions: () => [...waDeviceKeys.all, 'sessions'] as const,
  qr: (sessionId: string) => [...waDeviceKeys.all, 'qr', sessionId] as const,
  status: (sessionId: string) => [...waDeviceKeys.all, 'status', sessionId] as const,
}
