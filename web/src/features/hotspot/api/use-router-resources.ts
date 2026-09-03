import { useQuery } from '@tanstack/react-query'
import { networkClient } from '@/lib/api-client'
import { hotspotKeys } from './keys'

// Daftar nama parent queue statis router (/queue/simple print where dynamic=false).
export function useParentQueuesQuery(deviceId: string, enabled = true) {
  return useQuery({
    queryKey: hotspotKeys.parentQueues(deviceId),
    queryFn: async () => {
      const res = await networkClient.listParentQueues({ deviceId })
      return res.queues.map((q) => q.name)
    },
    enabled: enabled && Boolean(deviceId),
    staleTime: 60_000,
  })
}

// Daftar nama IP pool router (/ip/pool/print).
export function useIpPoolsQuery(deviceId: string, enabled = true) {
  return useQuery({
    queryKey: hotspotKeys.ipPools(deviceId),
    queryFn: async () => {
      const res = await networkClient.listIPPools({ deviceId })
      return res.pools.map((p) => p.name)
    },
    enabled: enabled && Boolean(deviceId),
    staleTime: 60_000,
  })
}
