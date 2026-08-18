export const pppKeys = {
  all: ['ppp'] as const,
  secrets: (deviceId?: string) => [...pppKeys.all, 'secrets', deviceId] as const,
  secretDetail: (deviceId?: string, id?: string) =>
    [...pppKeys.secrets(deviceId), id] as const,
  profiles: (deviceId?: string) => [...pppKeys.all, 'profiles', deviceId] as const,
  profileDetail: (deviceId?: string, id?: string) =>
    [...pppKeys.profiles(deviceId), id] as const,
  active: (deviceId?: string) => [...pppKeys.all, 'active', deviceId] as const,
  inactive: (deviceId?: string) => [...pppKeys.all, 'inactive', deviceId] as const,
}
