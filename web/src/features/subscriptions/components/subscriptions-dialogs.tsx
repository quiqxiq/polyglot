import { ChangePlanDialog } from './change-plan-dialog'
import { SuspendDialog } from './suspend-dialog'
import { LifecycleConfirmDialog } from './lifecycle-confirm-dialog'
import { useSubscriptions } from './subscriptions-provider'

type SubscriptionsDialogType = NonNullable<
  ReturnType<typeof useSubscriptions>['open']
>

const LIFECYCLE_DIALOGS: SubscriptionsDialogType[] = [
  'resume',
  'terminate',
  'activate',
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
      {/* Dialog dengan state form di-mount kondisional supaya state ter-reset
          otomatis setiap kali dibuka (tanpa setState di effect). */}
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
          action={open as 'resume' | 'terminate' | 'activate'}
          open={LIFECYCLE_DIALOGS.includes(open)}
          onOpenChange={() =>
            close(open as 'resume' | 'terminate' | 'activate')
          }
          currentRow={currentRow}
        />
      )}
    </>
  )
}
