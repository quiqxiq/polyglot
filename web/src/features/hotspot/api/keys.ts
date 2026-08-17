export const hotspotKeys = {
  all: ['hotspot'] as const,
  profiles: (deviceId: string) => [...hotspotKeys.all, 'profiles', deviceId] as const,
  users: (deviceId: string, profile?: string) => [...hotspotKeys.all, 'users', deviceId, profile || 'all'] as const,
  activeSessions: (deviceId: string) => [...hotspotKeys.all, 'active-sessions', deviceId] as const,
  dhcpLeases: (deviceId: string, mac?: string) => [...hotspotKeys.all, 'dhcp-leases', deviceId, mac || 'all'] as const,
}
