import { toast } from 'sonner'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { useHotspot } from '../../context/hotspot-context'
import { useDeleteHotspotIPBindingMutation } from '../../api/use-hotspot-bindings'
import { useDeviceStore } from '@/stores/device-store'

type BindingDeleteDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function BindingDeleteDialog({
  open,
  onOpenChange,
}: BindingDeleteDialogProps) {
  const { currentBinding, setCurrentBinding } = useHotspot()
  const { selectedDeviceId } = useDeviceStore()
  const deleteMutation = useDeleteHotspotIPBindingMutation()

  const handleDelete = async () => {
    if (!selectedDeviceId || !currentBinding) return

    onOpenChange(false)
    toast.promise(
      deleteMutation.mutateAsync({
        deviceId: selectedDeviceId,
        rosId: currentBinding.id,
      }),
      {
        loading: 'Deleting IP binding...',
        success: 'IP binding deleted successfully.',
        error: 'Failed to delete IP binding.',
      }
    )
    setCurrentBinding(null)
  }

  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onOpenChange}
      handleConfirm={handleDelete}
      title='Delete IP Binding'
      desc={
        <p>
          Are you sure you want to delete IP binding for{' '}
          <strong className='font-mono font-semibold'>
            {currentBinding?.macAddress || currentBinding?.address || 'this entry'}
          </strong>
          ? <br />
          This device may no longer be able to bypass the hotspot authentication.
        </p>
      }
      confirmText='Delete'
      destructive
      isLoading={deleteMutation.isPending}
    />
  )
}
