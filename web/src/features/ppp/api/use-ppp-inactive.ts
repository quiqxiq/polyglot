import { pppClient } from '@/lib/api-client'
import { ListPPPInactiveSecretsRequest } from '@/gen/v1/ppp_pb'
import { useQuery } from '@tanstack/react-query'
import { pppKeys } from './keys'

export function usePPPInactiveSecretsQuery(deviceId?: string) {
  return useQuery({
    queryKey: pppKeys.inactive(deviceId),
    queryFn: async () => {
      if (!deviceId) return []
      const res = await pppClient.listInactiveSecrets(
        new ListPPPInactiveSecretsRequest({ deviceId })
      )
      return res.secrets
    },
    enabled: !!deviceId,
  })
}
