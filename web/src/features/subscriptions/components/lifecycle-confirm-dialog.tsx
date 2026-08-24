import { toast } from 'sonner'
import { ConfirmDialog } from '@/components/confirm-dialog'
import {
  ActivateSubscriptionRequest,
  ResumeSubscriptionRequest,
  TerminateSubscriptionRequest,
  type Subscription,
} from '@/gen/v1/billing_pb'
import {
  useActivateSubscriptionMutation,
  useResumeSubscriptionMutation,
  useTerminateSubscriptionMutation,
} from '@/features/billing/api/use-billing'

type LifecycleAction = 'resume' | 'terminate' | 'activate'

const ACTION_META: Record<
  LifecycleAction,
  { title: string; desc: string; confirmText: string; successMsg: string; destructive: boolean }
> = {
  resume: {
    title: 'Resume Langganan',
    desc: 'Pulihkan langganan ini — user router akan diaktifkan kembali.',
    confirmText: 'Resume',
    successMsg: 'Langganan berhasil dipulihkan',
    destructive: false,
  },
  terminate: {
    title: 'Terminate Langganan',
    desc: 'Hentikan langganan ini secara permanen. User router akan dihapus dan status tidak bisa kembali aktif.',
    confirmText: 'Terminate',
    successMsg: 'Langganan berhasil dihentikan',
    destructive: true,
  },
  activate: {
    title: 'Activate Provisi',
    desc: 'Picu provisi ulang akun router untuk langganan ini.',
    confirmText: 'Activate',
    successMsg: 'Aktivasi provisi dikirim',
    destructive: false,
  },
}

interface LifecycleConfirmDialogProps {
  action: LifecycleAction
  open: boolean
  onOpenChange: (open: boolean) => void
  currentRow: Subscription | null
}

export function LifecycleConfirmDialog({
  action,
  open,
  onOpenChange,
  currentRow: subscription,
}: LifecycleConfirmDialogProps) {
  const resume = useResumeSubscriptionMutation()
  const terminate = useTerminateSubscriptionMutation()
  const activate = useActivateSubscriptionMutation()

  const meta = ACTION_META[action]
  const mutation =
    action === 'resume' ? resume : action === 'terminate' ? terminate : activate

  const handleConfirm = async () => {
    if (!subscription) return
    try {
      if (action === 'resume') {
        await resume.mutateAsync(
          new ResumeSubscriptionRequest({ subscriptionId: subscription.id })
        )
      } else if (action === 'terminate') {
        await terminate.mutateAsync(
          new TerminateSubscriptionRequest({ subscriptionId: subscription.id })
        )
      } else {
        await activate.mutateAsync(
          new ActivateSubscriptionRequest({ subscriptionId: subscription.id })
        )
      }
      toast.success(meta.successMsg)
      onOpenChange(false)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Aksi gagal, coba lagi')
    }
  }

  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onOpenChange}
      title={meta.title}
      desc={meta.desc}
      confirmText={mutation.isPending ? 'Memproses...' : meta.confirmText}
      destructive={meta.destructive}
      isLoading={mutation.isPending}
      handleConfirm={handleConfirm}
    />
  )
}
