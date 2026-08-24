export const reportsKeys = {
  all: ['reports'] as const,
  hotspot: (deviceId: string, day = '', month = '', year = '') =>
    [...reportsKeys.all, 'hotspot', deviceId, day, month, year] as const,
  list: (deviceId: string, day = '', month = '', year = '') =>
    [...reportsKeys.all, 'list', deviceId, day, month, year] as const,
  daily: (date = '') => [...reportsKeys.all, 'daily', date || 'today'] as const,
  monthly: (month = '') => [...reportsKeys.all, 'monthly', month || 'current'] as const,
  yearly: (year = 0) => [...reportsKeys.all, 'yearly', year || 'current'] as const,
}
