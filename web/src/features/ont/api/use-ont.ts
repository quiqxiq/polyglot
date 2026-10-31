import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  EnsureAcsPresetRequest,
  LinkOntRequest,
  PinAcsDeviceRequest,
  PushOntPppoeRequest,
  RebootOntRequest,
  RefreshOntTelemetryRequest,
  SetOntWifiRequest,
  SyncAcsDevicesRequest,
  UnlinkOntRequest,
  UnpinAcsDeviceRequest,
} from '@/gen/v1/ont_pb'
import { oltClient, ontClient } from '@/lib/api-client'
import { oltKeys } from '@/features/olt/api/keys'
import { ontKeys } from './keys'

export function useAcsConfigQuery() {
  return useQuery({
    queryKey: ontKeys.acsConfig(),
    queryFn: async () => {
      const res = await ontClient.getAcsConfig({})
      return res.config
    },
  })
}

export function useAcsDevicesQuery(onlyUnmatched: boolean, search: string) {
  return useQuery({
    queryKey: ontKeys.acsDevices(onlyUnmatched, search),
    queryFn: async () => {
      const res = await ontClient.listAcsDevices({
        onlyUnmatched,
        search,
        limit: 200,
      })
      return res.devices
    },
  })
}

/**
 * ONT terhubung ke pelanggan. Backend belum menyediakan filter customer pada
 * ListOnus, jadi seluruh inventori ONT difilter di sisi klien (MVP).
 */
export function useCustomerOnusQuery(customerId: string, enabled = true) {
  return useQuery({
    queryKey: ontKeys.customerOnus(customerId),
    queryFn: async () => {
      const res = await oltClient.listOnus({})
      return res.onus.filter((onu) => onu.customerId === customerId)
    },
    enabled: Boolean(customerId) && enabled,
  })
}

export function useUnlinkedOnusQuery(oltDeviceId: string) {
  return useQuery({
    queryKey: ontKeys.unlinkedOnus(oltDeviceId),
    queryFn: async () => {
      const res = await oltClient.listOnus({ oltDeviceId })
      return res.onus.filter((onu) => !onu.customerId && !onu.subscriptionId)
    },
    enabled: Boolean(oltDeviceId),
  })
}

export function useEnsureAcsPresetMutation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async () =>
      await ontClient.ensureAcsPreset(new EnsureAcsPresetRequest({})),
    onSettled: () => qc.invalidateQueries({ queryKey: ontKeys.all }),
  })
}

export function useSyncAcsDevicesMutation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async () =>
      await ontClient.syncAcsDevices(new SyncAcsDevicesRequest({})),
    onSettled: () => qc.invalidateQueries({ queryKey: ontKeys.all }),
  })
}

export function usePinAcsDeviceMutation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: { ontId: string; acsDeviceId: string }) =>
      await ontClient.pinAcsDevice(
        new PinAcsDeviceRequest({
          ontId: input.ontId,
          acsDeviceId: input.acsDeviceId,
        })
      ),
    onSettled: () => qc.invalidateQueries({ queryKey: ontKeys.all }),
  })
}

export function useUnpinAcsDeviceMutation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (ontId: string) =>
      await ontClient.unpinAcsDevice(new UnpinAcsDeviceRequest({ ontId })),
    onSettled: () => qc.invalidateQueries({ queryKey: ontKeys.all }),
  })
}

function invalidateOntData(qc: ReturnType<typeof useQueryClient>) {
  qc.invalidateQueries({ queryKey: ontKeys.all })
  qc.invalidateQueries({ queryKey: oltKeys.all })
}

export function useRefreshOntTelemetryMutation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (ontId: string) =>
      await ontClient.refreshOntTelemetry(
        new RefreshOntTelemetryRequest({ ontId })
      ),
    onSettled: () => invalidateOntData(qc),
  })
}

export function useRebootOntMutation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (ontId: string) =>
      await ontClient.rebootOnt(new RebootOntRequest({ ontId })),
    onSettled: () => invalidateOntData(qc),
  })
}

export function useSetOntWifiMutation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: {
      ontId: string
      ssid: string
      password: string
      band: string
    }) =>
      await ontClient.setOntWifi(
        new SetOntWifiRequest({
          ontId: input.ontId,
          ssid: input.ssid,
          password: input.password,
          band: input.band,
        })
      ),
    onSettled: () => invalidateOntData(qc),
  })
}

export function usePushOntPppoeMutation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: {
      ontId: string
      username: string
      password: string
    }) =>
      await ontClient.pushOntPppoe(
        new PushOntPppoeRequest({
          ontId: input.ontId,
          username: input.username,
          password: input.password,
        })
      ),
    onSettled: () => invalidateOntData(qc),
  })
}

export function useLinkOntMutation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: {
      ontId: string
      customerId: string
      subscriptionId?: string
      pppoeUsername?: string
    }) =>
      await ontClient.linkOnt(
        new LinkOntRequest({
          ontId: input.ontId,
          customerId: input.customerId,
          subscriptionId: input.subscriptionId ?? '',
          pppoeUsername: input.pppoeUsername ?? '',
        })
      ),
    onSettled: () => invalidateOntData(qc),
  })
}

export function useUnlinkOntMutation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (ontId: string) =>
      await ontClient.unlinkOnt(new UnlinkOntRequest({ ontId })),
    onSettled: () => invalidateOntData(qc),
  })
}

/** Pencarian ONU berdasarkan MAC/SN untuk form instalasi PSB. */
export function useFindOnusMutation() {
  return useMutation({
    mutationFn: async (input: { oltDeviceId: string; search: string }) => {
      const res = await oltClient.listOnus({
        oltDeviceId: input.oltDeviceId,
        search: input.search,
        limit: 50,
      })
      return res.onus
    },
  })
}
