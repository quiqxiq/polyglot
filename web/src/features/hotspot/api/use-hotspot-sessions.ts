import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { hotspotClient, networkClient } from '@/lib/api-client'
import { hotspotKeys } from './keys'
import {
  KickHotspotSessionRequest,
  RemoveHotspotHostRequest,
} from '@/gen/v1/hotspot_pb'
import { BlockDHCPLeaseRequest } from '@/gen/v1/network_pb'

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

export function useHotspotHostsQuery(deviceId: string, enabled = true) {
  return useQuery({
    queryKey: hotspotKeys.hosts(deviceId),
    queryFn: async () => {
      const res = await hotspotClient.listHosts({ deviceId })
      return res.hosts
    },
    enabled: Boolean(deviceId) && enabled,
  })
}

export function useHotspotServersQuery(deviceId: string, enabled = true) {
  return useQuery({
    queryKey: hotspotKeys.servers(deviceId),
    queryFn: async () => {
      const res = await hotspotClient.listHotspotServers({ deviceId })
      return res.servers
    },
    enabled: Boolean(deviceId) && enabled,
  })
}

export function useDHCPLeasesQuery(deviceId: string, macFilter = '', enabled = true) {
  return useQuery({
    queryKey: hotspotKeys.dhcpLeases(deviceId, macFilter),
    queryFn: async () => {
      const res = await networkClient.listDHCPLeases({ deviceId, macFilter })
      return res.leases
    },
    enabled: Boolean(deviceId) && enabled,
  })
}

export function useKickHotspotSessionMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: { deviceId: string; rosId: string }) => {
      return await hotspotClient.kickActiveSession(new KickHotspotSessionRequest(params))
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: hotspotKeys.activeSessions(variables.deviceId) })
    },
  })
}

export function useRemoveHotspotHostMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: { deviceId: string; rosId: string }) => {
      return await hotspotClient.removeHost(new RemoveHotspotHostRequest(params))
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: hotspotKeys.hosts(variables.deviceId) })
    },
  })
}

export function useBlockDHCPLeaseMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: { deviceId: string; rosId: string; blocked: boolean; comment?: string }) => {
      return await networkClient.blockDHCPLease(new BlockDHCPLeaseRequest(params))
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: hotspotKeys.dhcpLeases(variables.deviceId) })
    },
  })
}
