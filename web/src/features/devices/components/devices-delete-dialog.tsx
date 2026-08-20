'use client'

import { toast } from 'sonner'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { useDevicesContext } from './devices-provider'
import { useDeleteDeviceMutation } from '../api/use-devices'
import { DeleteDeviceRequest } from '@/gen/v1/device_pb'

export function DevicesDeleteDialog() {
  const { open, setOpen, currentRow, setCurrentRow } = useDevicesContext()
  const deleteMutation = useDeleteDeviceMutation()

  async function handleDelete() {
    if (!currentRow) return

    try {
      await deleteMutation.mutateAsync(new DeleteDeviceRequest({ id: currentRow.id }))
      toast.success(`Device "${currentRow.name}" deleted successfully`)
      setOpen(null)
      setCurrentRow(null)
    } catch (err: any) {
      toast.error(err.message || 'Failed to delete device')
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
      title='Delete Device'
      desc={
        <>
          Are you sure you want to delete <strong>{currentRow?.name}</strong> ({currentRow?.host})?
          This action cannot be undone.
        </>
      }
      confirmText='Delete'
      destructive
      isLoading={deleteMutation.isPending}
    />
  )
}
