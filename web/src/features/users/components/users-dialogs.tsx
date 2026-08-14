import { UsersActionDialog } from './users-action-dialog'
import { UsersDeleteDialog } from './users-delete-dialog'
import { useUsers } from './users-provider'
import { UsersResetDialog } from './users-reset-dialog'
import { UsersToggleDialog } from './users-toggle-dialog'

export function UsersDialogs() {
  const { open, setOpen, currentRow, setCurrentRow } = useUsers()

  const close = (v: boolean) => {
    if (!v) {
      setCurrentRow(null)
      setOpen(null)
    }
  }

  return (
    <>
      <UsersActionDialog
        currentRow={currentRow}
        open={open === 'create' || open === 'edit'}
        onOpenChange={close}
      />
      <UsersResetDialog
        user={currentRow}
        open={open === 'reset'}
        onOpenChange={close}
      />
      <UsersToggleDialog
        user={currentRow}
        open={open === 'toggle'}
        onOpenChange={close}
      />
      <UsersDeleteDialog
        user={currentRow}
        open={open === 'delete'}
        onOpenChange={close}
      />
    </>
  )
}
