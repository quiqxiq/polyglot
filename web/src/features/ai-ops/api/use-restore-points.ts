import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { RestoreMode } from '@/gen/v1/restore_point_pb'
import { toast } from 'sonner'
import { restorePointClient } from '@/lib/api-client'
import { aiOpsKeys } from './keys'

export interface RestorePointFilterInput {
  deviceId?: string
  status?: string
  limit?: number
  offset?: number
}

export function useRestorePointsQuery(filter: RestorePointFilterInput = {}) {
  return useQuery({
    queryKey: aiOpsKeys.restorePoints.list(filter as Record<string, unknown>),
    queryFn: async () => {
      const res = await restorePointClient.listRestorePoints({
        deviceId: filter.deviceId || '',
        status: filter.status || '',
        limit: filter.limit || 50,
        offset: filter.offset || 0,
      })
      return { restorePoints: res.restorePoints, total: res.total }
    },
  })
}

export function useRestorePointQuery(id: string) {
  return useQuery({
    queryKey: aiOpsKeys.restorePoints.detail(id),
    queryFn: async () =>
      (await restorePointClient.getRestorePoint({ id })).restorePoint,
    enabled: Boolean(id),
  })
}

export function useCreateRestorePointMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (deviceId: string) =>
      await restorePointClient.createRestorePoint({ deviceId }),
    onSuccess: () => {
      toast.success('Restore point dibuat')
      queryClient.invalidateQueries({ queryKey: aiOpsKeys.restorePoints.all() })
    },
    onError: (err: Error) =>
      toast.error(`Gagal membuat restore point: ${err.message}`),
  })
}

export function useRestoreFromPointMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (req: {
      id: string
      mode: RestoreMode
      confirmation: string
    }) => await restorePointClient.restoreFromPoint(req),
    onSuccess: () => {
      toast.success('Restore dijalankan')
      queryClient.invalidateQueries({ queryKey: aiOpsKeys.restorePoints.all() })
    },
    onError: (err: Error) => toast.error(`Gagal restore: ${err.message}`),
  })
}

export async function downloadRestoreArtifact(id: string, kind: string) {
  const res = await restorePointClient.downloadRestorePointArtifact({
    id,
    kind,
  })
  const blob = new Blob([res.payload as BlobPart], {
    type: res.contentType || 'application/octet-stream',
  })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = res.filename || `restore-${id}.bin`
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
}

export { RestoreMode }

export function useDeleteRestorePointMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) =>
      await restorePointClient.deleteRestorePoint({ id }),
    onSuccess: () => {
      toast.success('Restore point dihapus')
      queryClient.invalidateQueries({ queryKey: aiOpsKeys.restorePoints.all() })
    },
    onError: (err: Error) => toast.error(`Gagal menghapus: ${err.message}`),
  })
}
