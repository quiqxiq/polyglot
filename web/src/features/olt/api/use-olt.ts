import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  CreateOltRequest,
  DeleteOnuRequest,
  DeleteOltRequest,
  RebootOnuRequest,
  RenameOnuRequest,
  SaveOltConfigRequest,
  SetOnuAdminStateRequest,
  SyncOltRequest,
  TestOltConnectionRequest,
  UpdateOltRequest,
  type Onu,
} from '@/gen/v1/olt_pb'
import { oltClient } from '@/lib/api-client'
import { oltKeys } from './keys'

export interface OltFormInput {
  name: string
  host: string
  telnetPort: number
  sshPort: number
  snmpPort: number
  transport: string
  variant: string
  ponCount: number
  username: string
  password: string
  communityRo: string
  communityRw: string
  enabled: boolean
}

export function useOltsQuery() {
  return useQuery({
    queryKey: oltKeys.lists(),
    queryFn: async () => {
      const res = await oltClient.listOlts({})
      return res.olts
    },
  })
}

export function useOltQuery(id: string, enabled = true) {
  return useQuery({
    queryKey: oltKeys.detail(id),
    queryFn: async () => await oltClient.getOlt({ id }),
    enabled: Boolean(id) && enabled,
  })
}

export function useOnusQuery(oltDeviceId: string, ponIndex = 0, state = '') {
  return useQuery({
    queryKey: oltKeys.onus(oltDeviceId, ponIndex, state),
    queryFn: async () => {
      const res = await oltClient.listOnus({
        oltDeviceId,
        ponIndex,
        state,
      })
      return res.onus
    },
    enabled: Boolean(oltDeviceId),
  })
}

export function useOnuQuery(onuId: string | null) {
  return useQuery({
    queryKey: oltKeys.onu(onuId ?? ''),
    queryFn: async () => {
      const res = await oltClient.getOnu({ id: onuId ?? '' })
      return res.onu
    },
    enabled: Boolean(onuId),
  })
}

export function useOnuEventsQuery(onuId: string | null) {
  return useQuery({
    queryKey: oltKeys.onuEvents(onuId ?? ''),
    queryFn: async () => {
      const res = await oltClient.listOnuEvents({
        onuId: onuId ?? '',
        limit: 50,
      })
      return res.events
    },
    enabled: Boolean(onuId),
  })
}

export function useOnuOpticalHistoryQuery(onuId: string | null, hours: number) {
  return useQuery({
    queryKey: oltKeys.onuOpticalHistory(onuId ?? '', hours),
    queryFn: async () => {
      const res = await oltClient.getOnuOpticalHistory({
        onuId: onuId ?? '',
        hours,
      })
      return res.samples
    },
    enabled: Boolean(onuId),
  })
}

export function useCreateOltMutation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: OltFormInput) =>
      await oltClient.createOlt(
        new CreateOltRequest({
          name: input.name,
          host: input.host,
          telnetPort: input.telnetPort,
          sshPort: input.sshPort,
          snmpPort: input.snmpPort,
          transport: input.transport,
          variant: input.variant,
          ponCount: input.ponCount,
          username: input.username,
          password: input.password,
          communityRo: input.communityRo,
          communityRw: input.communityRw,
          enabled: input.enabled,
        })
      ),
    onSettled: () => qc.invalidateQueries({ queryKey: oltKeys.all }),
  })
}

export function useUpdateOltMutation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: OltFormInput & { id: string }) =>
      await oltClient.updateOlt(
        new UpdateOltRequest({
          id: input.id,
          name: input.name,
          host: input.host,
          telnetPort: input.telnetPort,
          sshPort: input.sshPort,
          snmpPort: input.snmpPort,
          transport: input.transport,
          variant: input.variant,
          ponCount: input.ponCount,
          username: input.username,
          password: input.password,
          communityRo: input.communityRo,
          communityRw: input.communityRw,
          enabled: input.enabled,
        })
      ),
    onSettled: () => qc.invalidateQueries({ queryKey: oltKeys.all }),
  })
}

export function useDeleteOltMutation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) =>
      await oltClient.deleteOlt(new DeleteOltRequest({ id })),
    onSettled: () => qc.invalidateQueries({ queryKey: oltKeys.all }),
  })
}

export function useTestOltConnectionMutation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) =>
      await oltClient.testOltConnection(new TestOltConnectionRequest({ id })),
    onSettled: (_data, _error, id) => {
      qc.invalidateQueries({ queryKey: oltKeys.lists() })
      qc.invalidateQueries({ queryKey: oltKeys.detail(id) })
    },
  })
}

export function useSyncOltMutation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (oltDeviceId: string) =>
      await oltClient.syncOlt(new SyncOltRequest({ oltDeviceId })),
    onSettled: () => qc.invalidateQueries({ queryKey: oltKeys.all }),
  })
}

export function useSaveOltConfigMutation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (oltDeviceId: string) =>
      await oltClient.saveOltConfig(new SaveOltConfigRequest({ oltDeviceId })),
    onSettled: () => qc.invalidateQueries({ queryKey: oltKeys.all }),
  })
}

function invalidateOnu(qc: ReturnType<typeof useQueryClient>, onu: Onu) {
  qc.invalidateQueries({ queryKey: oltKeys.onu(onu.id) })
  qc.invalidateQueries({ queryKey: oltKeys.detail(onu.oltDeviceId) })
  qc.invalidateQueries({ queryKey: oltKeys.lists() })
}

export function useRenameOnuMutation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: { onu: Onu; name: string }) =>
      await oltClient.renameOnu(
        new RenameOnuRequest({ onuId: input.onu.id, name: input.name })
      ),
    onSettled: (_data, _error, input) => invalidateOnu(qc, input.onu),
  })
}

export function useRebootOnuMutation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (onu: Onu) =>
      await oltClient.rebootOnu(new RebootOnuRequest({ onuId: onu.id })),
    onSettled: (_data, _error, onu) => invalidateOnu(qc, onu),
  })
}

export function useSetOnuAdminStateMutation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: { onu: Onu; enabled: boolean }) =>
      await oltClient.setOnuAdminState(
        new SetOnuAdminStateRequest({
          onuId: input.onu.id,
          enabled: input.enabled,
        })
      ),
    onSettled: (_data, _error, input) => invalidateOnu(qc, input.onu),
  })
}

export function useDeleteOnuMutation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (onu: Onu) =>
      await oltClient.deleteOnu(new DeleteOnuRequest({ onuId: onu.id })),
    onSettled: (_data, _error, onu) => invalidateOnu(qc, onu),
  })
}
