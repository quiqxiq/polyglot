import { toast } from 'sonner'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { useDeleteHotspotCookieMutation } from '../../api/use-hotspot-cookies'
import { useDeviceStore } from '@/stores/device-store'

type CookieClearAllDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function CookieClearAllDialog({
  open,
  onOpenChange,
}: CookieClearAllDialogProps) {
  const { selectedDeviceId } = useDeviceStore()
  const deleteMutation = useDeleteHotspotCookieMutation()

  const handleClearAll = async () => {
    if (!selectedDeviceId) return

    onOpenChange(false)
    toast.promise(
      deleteMutation.mutateAsync({
        deviceId: selectedDeviceId,
        rosId: 'all',
      }),
      {
        loading: 'Clearing all login cookies...',
        success: 'All hotspot cookies cleared successfully.',
        error: 'Failed to clear cookies.',
      }
    )
  }

  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onOpenChange}
      handleConfirm={handleClearAll}
      title='Clear All Hotspot Cookies'
      desc={
        <p>
          Are you sure you want to delete <strong>ALL</strong> hotspot login cookies on this router? <br />
          All currently remembered devices will be required to re-authenticate on the captive portal.
        </p>
      }
      confirmText='Clear All'
      destructive
      isLoading={deleteMutation.isPending}
    />
  )
}
