import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { deviceClient } from '@/lib/api-client'
import { deviceKeys } from './keys'
import type {
  UpdateDeviceRequest,
  DeleteDeviceRequest,
  TestDeviceConnectionRequest,
  CreateIsolationProfileRequest,
  UpdateIsolationProfileRequest,
  DeleteIsolationProfileRequest,
  ApplyRouterIntegrationScriptRequest,
} from '@/gen/v1/device_pb'


export function useDevicesQuery() {
  return useQuery({
    queryKey: deviceKeys.lists(),
    queryFn: async () => {
      const res = await deviceClient.listDevices({})
      return res.devices
    },
  })
}

export function useDeviceQuery(id: string, enabled = true) {
  return useQuery({
    queryKey: deviceKeys.detail(id),
    queryFn: async () => {
      const res = await deviceClient.getDevice({ id })
      return res.device
    },
    enabled: Boolean(id) && enabled,
  })
}

export function useUpdateDeviceMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: UpdateDeviceRequest) => {
      return await deviceClient.updateDevice(req)
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: deviceKeys.lists() })
      if (variables.device?.id) {
        queryClient.invalidateQueries({ queryKey: deviceKeys.detail(variables.device.id) })
      }
    },
  })
}

export function useDeleteDeviceMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: DeleteDeviceRequest) => {
      return await deviceClient.deleteDevice(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: deviceKeys.lists() })
    },
  })
}

export function useTestConnectionMutation() {
  return useMutation({
    mutationFn: async (req: TestDeviceConnectionRequest) => {
      return await deviceClient.testDeviceConnection(req)
    },
  })
}

export function useDevicePingConfigQuery(deviceId: string, enabled = true) {
  return useQuery({
    queryKey: [...deviceKeys.detail(deviceId), 'ping-config'],
    queryFn: async () => {
      return await deviceClient.getDevicePingConfig({ deviceId })
    },
    enabled: Boolean(deviceId) && enabled,
  })
}

export function useUpdateDevicePingConfigMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: { deviceId: string; enabled: boolean; target: string; retentionDays: number }) => {
      return await deviceClient.updateDevicePingConfig({
        deviceId: req.deviceId,
        config: {
          enabled: req.enabled,
          target: req.target,
          retentionDays: req.retentionDays,
        },
      })
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: deviceKeys.lists() })
      queryClient.invalidateQueries({ queryKey: deviceKeys.detail(variables.deviceId) })
      queryClient.invalidateQueries({ queryKey: [...deviceKeys.detail(variables.deviceId), 'ping-config'] })
    },
  })
}

export function useDevicePingMetricsQuery(
  req: {
    deviceId: string
    startTime?: string
    endTime?: string
    bucketInterval?: string
  },
  enabled = true
) {
  const hasRange = Boolean(req.startTime && req.endTime && req.startTime < req.endTime)

  return useQuery({
    queryKey: [
      ...deviceKeys.detail(req.deviceId),
      'ping-metrics',
      req.startTime,
      req.endTime,
      req.bucketInterval,
    ],
    queryFn: async () => {
      return await deviceClient.queryDevicePingMetrics({
        deviceId: req.deviceId,
        startTime: req.startTime,
        endTime: req.endTime,
        bucketInterval: req.bucketInterval,
      })
    },
    placeholderData: keepPreviousData,
    enabled: Boolean(req.deviceId) && hasRange && enabled,
    retry: 1,
  })
}

// ─── Profil Isolir & Integrasi Webhook Router ─────────────────────────────

export function useIsolationStatusQuery(deviceId: string, enabled = true) {
  return useQuery({
    queryKey: [...deviceKeys.detail(deviceId), 'isolation-status'],
    queryFn: async () => {
      return await deviceClient.getIsolationStatus({ deviceId })
    },
    enabled: Boolean(deviceId) && enabled,
    staleTime: 10_000,
  })
}

export function useCreateIsolationProfileMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: CreateIsolationProfileRequest) => {
      return await deviceClient.createIsolationProfile(req)
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: [...deviceKeys.detail(variables.deviceId), 'isolation-status'],
      })
    },
  })
}

export function useUpdateIsolationProfileMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: UpdateIsolationProfileRequest) => {
      return await deviceClient.updateIsolationProfile(req)
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: [...deviceKeys.detail(variables.deviceId), 'isolation-status'],
      })
    },
  })
}

export function useDeleteIsolationProfileMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: DeleteIsolationProfileRequest) => {
      return await deviceClient.deleteIsolationProfile(req)
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: [...deviceKeys.detail(variables.deviceId), 'isolation-status'],
      })
    },
  })
}

export function useRouterIntegrationScriptQuery(
  deviceId: string,
  serviceType = 'all',
  webhookUrl = '',
  enabled = true
) {
  return useQuery({
    queryKey: [...deviceKeys.detail(deviceId), 'integration-scripts', serviceType, webhookUrl],
    queryFn: async () => {
      return await deviceClient.getRouterIntegrationScript({
        deviceId,
        serviceType,
        webhookUrl,
      })
    },
    enabled: Boolean(deviceId) && enabled,
    staleTime: 60_000,
  })
}

export function useApplyRouterIntegrationScriptMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: ApplyRouterIntegrationScriptRequest) => {
      return await deviceClient.applyRouterIntegrationScript(req)
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: [...deviceKeys.detail(variables.deviceId), 'integration-scripts'],
      })
    },
  })
}
