import { useState } from 'react'
import { type Table } from '@tanstack/react-table'
import { Unplug } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { DataTableBulkActions as BulkActionsToolbar } from '@/components/data-table'
import type { PPPActiveSession } from '@/gen/v1/ppp_pb'
import { ActiveMultiKickDialog } from './active-multi-kick-dialog'

interface ActiveBulkActionsProps<TData> {
  table: Table<TData>
}

export function ActiveBulkActions<TData>({
  table,
}: ActiveBulkActionsProps<TData>) {
  const [showKickConfirm, setShowKickConfirm] = useState(false)

  const selectedRows = table.getFilteredSelectedRowModel().rows
  const selectedSessions = selectedRows.map((r) => r.original as PPPActiveSession)

  return (
    <>
      <BulkActionsToolbar table={table} entityName="session">
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="destructive"
              size="icon"
              className="h-8 w-8"
              onClick={() => setShowKickConfirm(true)}
            >
              <Unplug className="h-4 w-4" />
              <span className="sr-only">Disconnect Selected</span>
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            Disconnect Selected ({selectedSessions.length})
          </TooltipContent>
        </Tooltip>
      </BulkActionsToolbar>

      {showKickConfirm && (
        <ActiveMultiKickDialog
          open={showKickConfirm}
          onOpenChange={setShowKickConfirm}
          selectedSessions={selectedSessions}
          onSuccess={() => table.resetRowSelection()}
        />
      )}
    </>
  )
}
