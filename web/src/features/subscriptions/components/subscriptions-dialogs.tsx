import { ChangePlanDialog } from './change-plan-dialog'
import { SuspendDialog } from './suspend-dialog'
import { LifecycleConfirmDialog } from './lifecycle-confirm-dialog'
import { SubscriptionsCreateDialog } from './subscriptions-create-dialog'
import { SubscriptionsEditDialog } from './subscriptions-edit-dialog'
import { SubscriptionsDeleteDialog } from './subscriptions-delete-dialog'
import { useSubscriptions } from './subscriptions-provider'

type SubscriptionsDialogType = NonNullable<
  ReturnType<typeof useSubscriptions>['open']
>

const LIFECYCLE_DIALOGS: SubscriptionsDialogType[] = [
  'resume',
  'terminate',
  'activate',
  'isolate',
  'restore',
]

export function SubscriptionsDialogs() {
  const { open, setOpen, currentRow } = useSubscriptions()

  const close = (dialog: SubscriptionsDialogType) => {
    // useDialogState's setOpen toggles (same value → null), jadi selalu kirim
    // nilai dialog yang sedang aktif untuk menutupnya.
    setOpen(dialog)
  }

  return (
    <>
      {/* Create dialog tidak butuh currentRow — mount kondisional supaya
          state form ter-reset setiap kali dibuka. */}
      {open === 'create' && (
        <SubscriptionsCreateDialog
          open
          onOpenChange={() => close('create')}
        />
      )}
      {/* Dialog dengan state form di-mount kondisional supaya state ter-reset
          otomatis setiap kali dibuka (tanpa setState di effect). */}
      {currentRow && open === 'edit' && (
        <SubscriptionsEditDialog
          key={`edit-${currentRow.id}`}
          open={open === 'edit'}
          onOpenChange={() => close('edit')}
          currentRow={currentRow}
        />
      )}
      {currentRow && open === 'delete' && (
        <SubscriptionsDeleteDialog
          key={`delete-${currentRow.id}`}
          open={open === 'delete'}
          onOpenChange={() => close('delete')}
          currentRow={currentRow}
        />
      )}
      {currentRow && open === 'change-plan' && (
        <ChangePlanDialog
          key={`change-plan-${currentRow.id}`}
          open={open === 'change-plan'}
          onOpenChange={() => close('change-plan')}
          currentRow={currentRow}
        />
      )}
      {currentRow && open === 'suspend' && (
        <SuspendDialog
          key={`suspend-${currentRow.id}`}
          open={open === 'suspend'}
          onOpenChange={() => close('suspend')}
          currentRow={currentRow}
        />
      )}
      {currentRow && open !== null && LIFECYCLE_DIALOGS.includes(open) && (
        <LifecycleConfirmDialog
          action={open as 'resume' | 'terminate' | 'activate' | 'isolate' | 'restore'}
          open={LIFECYCLE_DIALOGS.includes(open)}
          onOpenChange={() =>
            close(open as 'resume' | 'terminate' | 'activate' | 'isolate' | 'restore')
          }
          currentRow={currentRow}
        />
      )}
    </>
  )
}

