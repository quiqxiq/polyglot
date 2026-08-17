export const hotspotKeys = {
  all: ['hotspot'] as const,
  profiles: (deviceId: string) => [...hotspotKeys.all, 'profiles', deviceId] as const,
  users: (deviceId: string, profile?: string, comment?: string) =>
    [...hotspotKeys.all, 'users', deviceId, profile || 'all', comment || 'all'] as const,
  user: (deviceId: string, rosId: string) =>
    [...hotspotKeys.all, 'user', deviceId, rosId] as const,
  activeSessions: (deviceId: string) =>
    [...hotspotKeys.all, 'active-sessions', deviceId] as const,
  hosts: (deviceId: string) =>
    [...hotspotKeys.all, 'hosts', deviceId] as const,
  servers: (deviceId: string) =>
    [...hotspotKeys.all, 'servers', deviceId] as const,
  dhcpLeases: (deviceId: string, mac?: string) =>
    [...hotspotKeys.all, 'dhcp-leases', deviceId, mac || 'all'] as const,
  voucherBatch: (deviceId: string, comment: string) =>
    [...hotspotKeys.all, 'voucher-batch', deviceId, comment] as const,
  expireMonitorStatus: (deviceId: string) =>
    [...hotspotKeys.all, 'expire-monitor', deviceId] as const,
  templates: () => [...hotspotKeys.all, 'templates'] as const,
  templateSection: (templateName: string, section: string) =>
    [...hotspotKeys.all, 'template-section', templateName, section] as const,
}
