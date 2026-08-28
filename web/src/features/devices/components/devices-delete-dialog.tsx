'use client'

import { toast } from 'sonner'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { useDevicesContext } from './devices-provider'
import { useDeleteDeviceMutation } from '../api/use-devices'
import { DeleteDeviceRequest } from '@/gen/v1/device_pb'
import { getErrorMessage } from '../lib/formatters'

export function DevicesDeleteDialog() {
  const { open, setOpen, currentRow, setCurrentRow } = useDevicesContext()
  const deleteMutation = useDeleteDeviceMutation()

  async function handleDelete() {
    if (!currentRow) return

    try {
      await deleteMutation.mutateAsync(new DeleteDeviceRequest({ id: currentRow.id }))
      toast.success(`Perangkat "${currentRow.name}" berhasil dihapus`)
      setOpen(null)
      setCurrentRow(null)
    } catch (err: unknown) {
      toast.error(getErrorMessage(err, 'Gagal menghapus perangkat'))
    }
  }

  return (
    <ConfirmDialog
      open={open === 'delete'}
      onOpenChange={() => {
        setOpen(null)
        setCurrentRow(null)
      }}
      handleConfirm={handleDelete}
      title='Hapus Perangkat'
      desc={
        <>
          Apakah Anda yakin ingin menghapus <strong>{currentRow?.name}</strong> ({currentRow?.host})?
          Tindakan ini tidak dapat dibatalkan.
        </>
      }
      confirmText='Hapus'
      destructive
      isLoading={deleteMutation.isPending}
    />
  )
}
