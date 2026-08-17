import { useState } from 'react'
import { type Table } from '@tanstack/react-table'
import { Trash2, RotateCcw, Printer } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { DataTableBulkActions as BulkActionsToolbar } from '@/components/data-table'
import { useResetHotspotUserCountersMutation } from '../../api/use-hotspot-users'
import { useDeviceStore } from '@/stores/device-store'
import { useHotspot } from '../../context/hotspot-context'
import type { HotspotUser } from '@/gen/v1/hotspot_pb'
import { UsersMultiDeleteDialog } from './users-multi-delete-dialog'

type UsersBulkActionsProps<TData> = {
  table: Table<TData>
}

export function UsersBulkActions<TData>({
  table,
}: UsersBulkActionsProps<TData>) {
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const { selectedDeviceId } = useDeviceStore()
  const { setOpen, setPrintBatchComment, setPrintSingleUserId } = useHotspot()
  const resetMutation = useResetHotspotUserCountersMutation()

  const selectedRows = table.getFilteredSelectedRowModel().rows
  const selectedUsers = selectedRows.map((r) => r.original as HotspotUser)

  const handleBulkResetCounters = async () => {
    if (!selectedDeviceId || selectedUsers.length === 0) return

    toast.promise(
      (async () => {
        for (const u of selectedUsers) {
          await resetMutation.mutateAsync({
            deviceId: selectedDeviceId,
            rosId: u.id,
          })
        }
        table.resetRowSelection()
      })(),
      {
        loading: `Resetting counters for ${selectedUsers.length} user${selectedUsers.length > 1 ? 's' : ''}...`,
        success: `Counters reset for ${selectedUsers.length} user${selectedUsers.length > 1 ? 's' : ''}.`,
        error: 'Failed to reset counters for some users.',
      }
    )
  }

  const handleBulkPrint = () => {
    if (selectedUsers.length === 0) return

    if (selectedUsers.length === 1) {
      setPrintSingleUserId(selectedUsers[0].id)
      setPrintBatchComment('')
      setOpen('voucher-print')
      return
    }

    // Check if all selected users share a common comment/batch tag
    const comments = Array.from(
      new Set(selectedUsers.map((u) => u.comment?.trim()).filter(Boolean))
    )
    if (comments.length === 1 && comments[0]) {
      setPrintBatchComment(comments[0])
      setPrintSingleUserId('')
      setOpen('voucher-print')
    } else {
      // If multiple different comments, notify user or print first
      setPrintBatchComment(comments[0] || '')
      setPrintSingleUserId('')
      setOpen('voucher-print')
    }
  }

  return (
    <>
      <BulkActionsToolbar table={table} entityName='user'>
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant='outline'
              size='icon'
              onClick={handleBulkPrint}
              className='size-8'
              aria-label='Print vouchers for selected users'
            >
              <Printer className='size-4 text-primary' />
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p>Print vouchers</p>
          </TooltipContent>
        </Tooltip>

        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant='outline'
              size='icon'
              onClick={handleBulkResetCounters}
              className='size-8'
              aria-label='Reset counters for selected users'
            >
              <RotateCcw className='size-4 text-amber-500' />
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p>Reset counters</p>
          </TooltipContent>
        </Tooltip>

        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant='destructive'
              size='icon'
              onClick={() => setShowDeleteConfirm(true)}
              className='size-8'
              aria-label='Delete selected users'
            >
              <Trash2 className='size-4' />
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p>Delete users</p>
          </TooltipContent>
        </Tooltip>
      </BulkActionsToolbar>

      <UsersMultiDeleteDialog
        open={showDeleteConfirm}
        onOpenChange={setShowDeleteConfirm}
        table={table}
      />
    </>
  )
}
