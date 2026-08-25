import {
  ArrowUpCircle,
  Ban,
  MoreHorizontal,
  PauseCircle,
  Pencil,
  PlayCircle,
  Repeat,
  Trash2,
} from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Button } from '@/components/ui/button'
import type { Subscription } from '@/gen/v1/billing_pb'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'
import { useSubscriptions } from './subscriptions-provider'

interface SubscriptionsRowActionsProps {
  subscription: Subscription
}

export function SubscriptionsRowActions({ subscription }: SubscriptionsRowActionsProps) {
  const { setOpen, setCurrentRow } = useSubscriptions()
  const permissions = useAuthStore((s) => s.auth.user?.permissions)
  const canManage = canPermission(permissions, 'billing:manage')

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
      <DropdownMenuContent align='end' className='w-48'>
        {status === 'ACTIVE' && (
          <>
            <DropdownMenuItem onClick={() => act('change-plan')}>
              <Repeat className='mr-2 h-4 w-4' />
              Change Plan
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={() => act('suspend')}>
              <PauseCircle className='mr-2 h-4 w-4' />
              Suspend
            </DropdownMenuItem>
          </>
        )}
        {status === 'ISOLATED' && (
          <>
            <DropdownMenuItem onClick={() => act('resume')}>
              <PlayCircle className='mr-2 h-4 w-4' />
              Resume (Pulihkan)
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
