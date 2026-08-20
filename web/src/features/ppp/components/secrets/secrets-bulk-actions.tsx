import { useState } from 'react'
import { type Table } from '@tanstack/react-table'
import { CheckCircle2, PowerOff, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { DataTableBulkActions as BulkActionsToolbar } from '@/components/data-table'
import { useBulkSetPPPSecretsDisabledMutation } from '../../api/use-ppp-secrets'
import { useDeviceStore } from '@/stores/device-store'
import type { PPPSecret } from '@/gen/v1/ppp_pb'
import { SecretsMultiDeleteDialog } from './secrets-multi-delete-dialog'

interface SecretsBulkActionsProps<TData> {
  table: Table<TData>
}

export function SecretsBulkActions<TData>({
  table,
}: SecretsBulkActionsProps<TData>) {
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const selectedDeviceId = useDeviceStore((state) => state.selectedDeviceId)
  const bulkSetDisabledMutation = useBulkSetPPPSecretsDisabledMutation()

  const selectedRows = table.getFilteredSelectedRowModel().rows
  const selectedSecrets = selectedRows.map((r) => r.original as PPPSecret)

  const handleBulkToggleDisabled = (disabled: boolean) => {
    if (!selectedDeviceId || selectedSecrets.length === 0) return

    toast.promise(
      bulkSetDisabledMutation.mutateAsync({
        deviceId: selectedDeviceId,
        rosIds: selectedSecrets.map((s) => s.id),
        disabled,
      }),
      {
        loading: `${disabled ? 'Disabling' : 'Enabling'} ${selectedSecrets.length} secret${selectedSecrets.length > 1 ? 's' : ''}...`,
        success: () => {
          table.resetRowSelection()
          return `${selectedSecrets.length} secret${selectedSecrets.length > 1 ? 's' : ''} ${disabled ? 'disabled' : 'enabled'}.`
        },
        error: 'Failed to update secrets status.',
      }
    )
  }

  return (
    <>
      <BulkActionsToolbar table={table} entityName="secret">
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="outline"
              size="icon"
              className="h-8 w-8 text-emerald-500 hover:text-emerald-600 hover:bg-emerald-500/10"
              onClick={() => handleBulkToggleDisabled(false)}
            >
              <CheckCircle2 className="h-4 w-4" />
              <span className="sr-only">Enable Selected</span>
            </Button>
          </TooltipTrigger>
          <TooltipContent>Enable Selected ({selectedSecrets.length})</TooltipContent>
        </Tooltip>

        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="outline"
              size="icon"
              className="h-8 w-8 text-amber-500 hover:text-amber-600 hover:bg-amber-500/10"
              onClick={() => handleBulkToggleDisabled(true)}
            >
              <PowerOff className="h-4 w-4" />
              <span className="sr-only">Disable Selected</span>
            </Button>
          </TooltipTrigger>
          <TooltipContent>Disable Selected ({selectedSecrets.length})</TooltipContent>
        </Tooltip>

        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="destructive"
              size="icon"
              className="h-8 w-8"
              onClick={() => setShowDeleteConfirm(true)}
            >
              <Trash2 className="h-4 w-4" />
              <span className="sr-only">Delete Selected</span>
            </Button>
          </TooltipTrigger>
          <TooltipContent>Delete Selected ({selectedSecrets.length})</TooltipContent>
        </Tooltip>
      </BulkActionsToolbar>

      {showDeleteConfirm && (
        <SecretsMultiDeleteDialog
          open={showDeleteConfirm}
          onOpenChange={setShowDeleteConfirm}
          selectedSecrets={selectedSecrets}
          onSuccess={() => table.resetRowSelection()}
        />
      )}
    </>
  )
}
