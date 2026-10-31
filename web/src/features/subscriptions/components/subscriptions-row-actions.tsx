import {
  ArrowUpCircle,
  Ban,
  MoreHorizontal,
  PauseCircle,
  Pencil,
  PlayCircle,
  RefreshCw,
  Repeat,
  ShieldAlert,
  ShieldCheck,
  Trash2,
} from 'lucide-react'
import { toast } from 'sonner'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Button } from '@/components/ui/button'
import type { Subscription } from '@/gen/v1/subscription_pb'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'
import { useRetryProvisionMutation } from '@/features/billing/api/use-billing'
import { useSubscriptions } from './subscriptions-provider'

interface SubscriptionsRowActionsProps {
  subscription: Subscription
}

export function SubscriptionsRowActions({ subscription }: SubscriptionsRowActionsProps) {
  const { setOpen, setCurrentRow } = useSubscriptions()
  const permissions = useAuthStore((s) => s.auth.user?.permissions)
  const canManage = canPermission(permissions, 'billing:manage')
  const retryProvision = useRetryProvisionMutation()
  const provisionFailed =
    subscription.provisionStatus === 'FAILED' ||
    subscription.provisionStatus === 'PENDING'

  const handleRetryProvision = async () => {
    try {
      const res = await retryProvision.mutateAsync({ id: subscription.id })
      toast.success(
        `Job provisioning diantrekan (${res.jobId}) — status ${res.status}`
      )
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : 'Gagal mengantrekan ulang provisi'
      )
    }
  }

  if (!canManage) return null

  const act = (
    dialog:
      | 'create'
      | 'edit'
      | 'delete'
      | 'change-plan'
      | 'suspend'
      | 'resume'
      | 'terminate'
      | 'activate'
      | 'isolate'
      | 'restore'
  ) => {
    setCurrentRow(subscription)
    setOpen(dialog)
  }

  const status = subscription.status
  // Activate hanya relevan saat provision belum sukses.
  const canActivate = subscription.provisionStatus !== 'OK'

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant='ghost' className='h-8 w-8 p-0'>
          <span className='sr-only'>Open menu</span>
          <MoreHorizontal className='h-4 w-4' />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end' className='w-52'>
        {provisionFailed && (
          <>
            <DropdownMenuItem
              onClick={handleRetryProvision}
              disabled={retryProvision.isPending}
              className='text-blue-600 focus:text-blue-600 dark:text-blue-400'
            >
              <RefreshCw className='mr-2 h-4 w-4' />
              Retry Provision
            </DropdownMenuItem>
            <DropdownMenuSeparator />
          </>
        )}
        {status === 'ACTIVE' && (
          <>
            <DropdownMenuItem onClick={() => act('change-plan')}>
              <Repeat className='mr-2 h-4 w-4' />
              Change Plan
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={() => act('isolate')}
              className='text-amber-600 focus:text-amber-600 dark:text-amber-400'
            >
              <ShieldAlert className='mr-2 h-4 w-4' />
              Isolir Pelanggan
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => act('suspend')}>
              <PauseCircle className='mr-2 h-4 w-4' />
              Suspend
            </DropdownMenuItem>
          </>
        )}
        {status === 'ISOLATED' && (
          <>
            <DropdownMenuItem
              onClick={() => act('restore')}
              className='text-emerald-600 focus:text-emerald-600 font-semibold dark:text-emerald-400'
            >
              <ShieldCheck className='mr-2 h-4 w-4' />
              Buka Isolir (Restore)
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => act('suspend')}>
              <PauseCircle className='mr-2 h-4 w-4' />
              Suspend
            </DropdownMenuItem>
          </>
        )}
        {status === 'SUSPENDED' && (
          <DropdownMenuItem onClick={() => act('resume')}>
            <PlayCircle className='mr-2 h-4 w-4' />
            Resume
          </DropdownMenuItem>
        )}
        {canActivate && (
          <>
            <DropdownMenuItem onClick={() => act('activate')}>
              <ArrowUpCircle className='mr-2 h-4 w-4' />
              Activate
            </DropdownMenuItem>
            <DropdownMenuSeparator />
          </>
        )}
        {status !== 'TERMINATED' && (
          <DropdownMenuItem
            onClick={() => act('terminate')}
            className='text-destructive focus:text-destructive'
          >
            <Ban className='mr-2 h-4 w-4' />
            Terminate
          </DropdownMenuItem>
        )}
        {/* CRUD tersedia untuk semua status. */}
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={() => act('edit')}>
          <Pencil className='mr-2 h-4 w-4' />
          Edit
        </DropdownMenuItem>
        <DropdownMenuItem
          onClick={() => act('delete')}
          className='text-destructive focus:text-destructive'
        >
          <Trash2 className='mr-2 h-4 w-4' />
          Delete
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
