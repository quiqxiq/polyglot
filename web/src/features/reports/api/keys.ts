export const reportsKeys = {
  all: ['reports'] as const,
  list: (deviceId: string, day = '', month = '', year = '') =>
    [...reportsKeys.all, 'list', deviceId, day, month, year] as const,
}
