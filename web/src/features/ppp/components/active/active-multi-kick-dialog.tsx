import { ConfirmDialog } from '@/components/confirm-dialog'
import { useDeviceStore } from '@/stores/device-store'
import type { PPPActiveSession } from '@/gen/v1/ppp_pb'
import { useBulkKickPPPActiveSessionsMutation } from '../../api/use-ppp-active'

interface ActiveMultiKickDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  selectedSessions: PPPActiveSession[]
  onSuccess?: () => void
}

export function ActiveMultiKickDialog({
  open,
  onOpenChange,
  selectedSessions,
  onSuccess,
}: ActiveMultiKickDialogProps) {
  const selectedDeviceId = useDeviceStore((state) => state.selectedDeviceId)
  const bulkKickMutation = useBulkKickPPPActiveSessionsMutation()

  const handleKick = async () => {
    if (!selectedDeviceId || selectedSessions.length === 0) return
    await bulkKickMutation.mutateAsync({
      deviceId: selectedDeviceId,
      rosIds: selectedSessions.map((s) => s.id),
    })
    onOpenChange(false)
    onSuccess?.()
  }

  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onOpenChange}
      handleConfirm={handleKick}
      disabled={bulkKickMutation.isPending}
      title={`Disconnect ${selectedSessions.length} Active Sessions`}
      desc={
        <span>
          Are you sure you want to forcibly disconnect{' '}
          <strong>{selectedSessions.length}</strong> selected active subscriber session(s)?
          Their CPE connections will be terminated immediately.
        </span>
      }
      confirmText={`Disconnect ${selectedSessions.length} Sessions`}
      destructive
    />
  )
}
