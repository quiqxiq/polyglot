import { toast } from 'sonner'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { useHotspot } from '../../context/hotspot-context'
import { useDeleteHotspotCookieMutation } from '../../api/use-hotspot-cookies'
import { useDeviceStore } from '@/stores/device-store'

type CookieDeleteDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function CookieDeleteDialog({
  open,
  onOpenChange,
}: CookieDeleteDialogProps) {
  const { currentCookie, setCurrentCookie } = useHotspot()
  const { selectedDeviceId } = useDeviceStore()
  const deleteMutation = useDeleteHotspotCookieMutation()

  const handleDelete = async () => {
    if (!selectedDeviceId || !currentCookie) return

    onOpenChange(false)
    toast.promise(
      deleteMutation.mutateAsync({
        deviceId: selectedDeviceId,
        rosId: currentCookie.id,
      }),
      {
        loading: 'Deleting login cookie...',
        success: 'Cookie deleted successfully.',
        error: 'Failed to delete cookie.',
      }
    )
    setCurrentCookie(null)
  }

  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onOpenChange}
      handleConfirm={handleDelete}
      title='Remove Hotspot Cookie'
      desc={
        <p>
          Are you sure you want to remove the login cookie for user{' '}
          <strong className='font-semibold'>{currentCookie?.user}</strong> (MAC:{' '}
          <span className='font-mono'>{currentCookie?.macAddress}</span>)? <br />
          This user will be forced to log in again on their next connection.
        </p>
      }
      confirmText='Remove'
      destructive
      isLoading={deleteMutation.isPending}
    />
  )
}
