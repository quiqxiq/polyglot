import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { hotspotClient } from '@/lib/api-client'
import { hotspotKeys } from './keys'
import type {
  GenerateVouchersRequest,
  KickHotspotSessionRequest,
  BlockDHCPLeaseRequest,
} from '@/gen/v1/hotspot_pb'

export function useHotspotProfilesQuery(deviceId: string, enabled = true) {
  return useQuery({
    queryKey: hotspotKeys.profiles(deviceId),
    queryFn: async () => {
      const res = await hotspotClient.listProfiles({ deviceId })
      return res.profiles
    },
    enabled: Boolean(deviceId) && enabled,
  })
}

export function useHotspotUsersQuery(deviceId: string, profile = '', enabled = true) {
  return useQuery({
    queryKey: hotspotKeys.users(deviceId, profile),
    queryFn: async () => {
      const res = await hotspotClient.listUsers({ deviceId, profile })
      return res.users
    },
    enabled: Boolean(deviceId) && enabled,
  })
}

export function useHotspotActiveSessionsQuery(deviceId: string, enabled = true) {
  return useQuery({
    queryKey: hotspotKeys.activeSessions(deviceId),
    queryFn: async () => {
      const res = await hotspotClient.listActiveSessions({ deviceId })
      return res.sessions
    },
    enabled: Boolean(deviceId) && enabled,
  })
}

export function useDHCPLeasesQuery(deviceId: string, macFilter = '', enabled = true) {
  return useQuery({
    queryKey: hotspotKeys.dhcpLeases(deviceId, macFilter),
    queryFn: async () => {
      const res = await hotspotClient.listDHCPLeases({ deviceId, macFilter })
      return res.leases
    },
    enabled: Boolean(deviceId) && enabled,
  })
}

export function useGenerateVouchersMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: GenerateVouchersRequest) => {
      return await hotspotClient.generateVouchers(req)
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: hotspotKeys.users(variables.deviceId) })
    },
  })
}

export function useKickHotspotSessionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: KickHotspotSessionRequest) => {
      return await hotspotClient.kickActiveSession(req)
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: hotspotKeys.activeSessions(variables.deviceId) })
    },
  })
}

export function useBlockDHCPLeaseMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: BlockDHCPLeaseRequest) => {
      return await hotspotClient.blockDHCPLease(req)
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: hotspotKeys.dhcpLeases(variables.deviceId) })
    },
  })
}
