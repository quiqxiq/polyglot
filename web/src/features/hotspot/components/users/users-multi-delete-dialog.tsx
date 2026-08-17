import { useState } from 'react'
import { type Table } from '@tanstack/react-table'
import { AlertTriangle } from 'lucide-react'
import { toast } from 'sonner'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { useDeleteHotspotUserMutation } from '../../api/use-hotspot-users'
import { useDeviceStore } from '@/stores/device-store'
import type { HotspotUser } from '@/gen/v1/hotspot_pb'

type UsersMultiDeleteDialogProps<TData> = {
  open: boolean
  onOpenChange: (open: boolean) => void
  table: Table<TData>
}

const CONFIRM_WORD = 'DELETE'

export function UsersMultiDeleteDialog<TData>({
  open,
  onOpenChange,
  table,
}: UsersMultiDeleteDialogProps<TData>) {
  const [value, setValue] = useState('')
  const { selectedDeviceId } = useDeviceStore()
  const deleteMutation = useDeleteHotspotUserMutation()

  const selectedRows = table.getFilteredSelectedRowModel().rows

  const handleDelete = async () => {
    if (value.trim() !== CONFIRM_WORD) {
      toast.error(`Please type "${CONFIRM_WORD}" to confirm.`)
      return
    }

    if (!selectedDeviceId) return

    onOpenChange(false)
    const usersToDelete = selectedRows.map((r) => r.original as HotspotUser)

    toast.promise(
      (async () => {
        for (const user of usersToDelete) {
          await deleteMutation.mutateAsync({
            deviceId: selectedDeviceId,
            rosId: user.id,
          })
        }
        setValue('')
        table.resetRowSelection()
      })(),
      {
        loading: `Deleting ${usersToDelete.length} user${usersToDelete.length > 1 ? 's' : ''}...`,
        success: `Deleted ${usersToDelete.length} user${usersToDelete.length > 1 ? 's' : ''} successfully.`,
        error: 'Failed to delete some or all selected users.',
      }
    )
  }

  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onOpenChange}
      form='users-multi-delete-form'
      disabled={value.trim() !== CONFIRM_WORD}
      title={
        <span className='text-destructive'>
          <AlertTriangle
            className='me-1 inline-block stroke-destructive'
            size={18}
          />{' '}
          Delete {selectedRows.length}{' '}
          {selectedRows.length > 1 ? 'users' : 'user'}
        </span>
      }
      desc={
        <form
          id='users-multi-delete-form'
          onSubmit={(e) => {
            e.preventDefault()
            handleDelete()
          }}
          className='space-y-4'
        >
          <p className='mb-2'>
            Are you sure you want to permanently delete the selected users from
            MikroTik Hotspot? <br />
            This action cannot be undone.
          </p>

          <Label className='my-4 flex flex-col items-start gap-1.5'>
            <span>Confirm by typing "{CONFIRM_WORD}":</span>
            <Input
              value={value}
              onChange={(e) => setValue(e.target.value)}
              placeholder={`Type "${CONFIRM_WORD}" to confirm.`}
              autoFocus
            />
          </Label>

          <Alert variant='destructive'>
            <AlertTitle>Warning!</AlertTitle>
            <AlertDescription>
              Please be careful, this operation directly removes users from the
              active router.
            </AlertDescription>
          </Alert>
        </form>
      }
      confirmText='Delete'
      destructive
    />
  )
}
