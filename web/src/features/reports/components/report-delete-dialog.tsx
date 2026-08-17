import { toast } from 'sonner'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { useDeleteHotspotReportMutation } from '../api/use-reports'
import { useDeviceStore } from '@/stores/device-store'
import { useReports } from '../context/reports-context'

export function ReportDeleteDialog() {
  const { open, setOpen, currentReport, setCurrentReport } = useReports()
  const { selectedDeviceId } = useDeviceStore()
  const deleteMutation = useDeleteHotspotReportMutation()

  const isOpen = open === 'report-delete'

  const handleConfirm = async () => {
    if (!currentReport) return
    try {
      await deleteMutation.mutateAsync({
        deviceId: selectedDeviceId,
        rosId: currentReport.id,
      })
      toast.success('Sales report transaction entry deleted!')
      setOpen(null)
      setCurrentReport(null)
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Failed to delete report entry')
    }
  }

  return (
    <ConfirmDialog
      open={isOpen}
      onOpenChange={(val) => {
        if (!val) {
          setOpen(null)
          setCurrentReport(null)
        }
      }}
      destructive
      title={`Delete Transaction: ${currentReport?.username}?`}
      desc={
        <>
          Are you sure you want to delete sales record for voucher{' '}
          <strong>{currentReport?.username}</strong> ({currentReport?.date} {currentReport?.time})?
        </>
      }
      confirmText={deleteMutation.isPending ? 'Deleting...' : 'Delete Transaction'}
      handleConfirm={handleConfirm}
    />
  )
}
