import { type ColumnDef } from '@tanstack/react-table'
import { type User } from '@/gen/v1/users_pb'
import { MoreHorizontal, ShieldCheck, ShieldOff } from 'lucide-react'
import { useAuthStore } from '@/stores/auth-store'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { roleClassName, roleLabel } from '../data/roles'
import { useUsers } from './users-provider'

function formatDate(unixSeconds: bigint): string {
  if (!unixSeconds) return '-'
  const d = new Date(Number(unixSeconds) * 1000)
  return d.toLocaleDateString(undefined, {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}

export const usersColumns: ColumnDef<User>[] = [
  {
    accessorKey: 'username',
    header: 'User',
    cell: ({ row }) => {
      const { username, fullName, email } = row.original
      return (
        <div className='ps-1'>
          <div className='font-medium'>{fullName || username}</div>
          <div className='text-xs text-muted-foreground'>
            {fullName ? `@${username} • ` : ''}{email}
          </div>
        </div>
      )
    },
    enableHiding: false,
  },
  {
    id: 'contact',
    header: 'Kontak & Spesialisasi',
    cell: ({ row }) => {
      const { phoneNumber, specialization } = row.original
      if (!phoneNumber && !specialization) {
        return <span className='text-xs text-muted-foreground'>-</span>
      }
      return (
        <div className='text-xs space-y-0.5'>
          {phoneNumber && <div className='font-mono font-medium'>{phoneNumber}</div>}
          {specialization && <div className='text-muted-foreground'>{specialization}</div>}
        </div>
      )
    },
  },
  {
    id: 'roles',
    header: 'Role',
    cell: ({ row }) => {
      const { role, roles } = row.original
      const extras = (roles ?? []).filter((r) => r !== role)
      return (
        <div className='flex flex-wrap items-center gap-1'>
          <Badge
            variant='outline'
            className={cn('capitalize', roleClassName(role))}
          >
            {roleLabel(role)}
          </Badge>
          {extras.slice(0, 2).map((r) => (
            <Badge
              key={r}
              variant='outline'
              className={cn('capitalize', roleClassName(r))}
            >
              {roleLabel(r)}
            </Badge>
          ))}
          {extras.length > 2 && (
            <Badge variant='secondary'>+{extras.length - 2}</Badge>
          )}
        </div>
      )
    },
    enableSorting: false,
  },
  {
    id: 'assignedDevices',
    header: 'Assigned Routers',
    cell: ({ row }) => {
      const { role, assignedDeviceIds } = row.original
      if (role === 'owner') {
        return (
          <Badge
            variant='outline'
            className='bg-purple-500/10 text-purple-700 dark:text-purple-400 border-purple-200 text-xs'
          >
            All Routers (Global)
          </Badge>
        )
      }
      if (!assignedDeviceIds || assignedDeviceIds.length === 0) {
        return <span className='text-xs text-muted-foreground italic'>None (No Access)</span>
      }
      return (
        <div className='flex flex-wrap items-center gap-1'>
          {assignedDeviceIds.slice(0, 2).map((id) => (
            <Badge key={id} variant='secondary' className='text-xs font-mono font-normal'>
              {id}
            </Badge>
          ))}
          {assignedDeviceIds.length > 2 && (
            <Badge variant='outline' className='text-xs'>
              +{assignedDeviceIds.length - 2} more
            </Badge>
          )}
        </div>
      )
    },
    enableSorting: false,
  },
  {
    id: 'status',
    header: 'Status',
    cell: ({ row }) =>
      row.original.isActive ? (
        <Badge className='bg-emerald-500/15 text-emerald-700 dark:text-emerald-400'>
          Active
        </Badge>
      ) : (
        <Badge
          variant='outline'
          className='bg-zinc-500/10 text-zinc-600 dark:text-zinc-400'
        >
          Disabled
        </Badge>
      ),
    enableSorting: false,
  },
  {
    id: 'created',
    header: 'Created',
    cell: ({ row }) => (
      <div className='text-sm text-muted-foreground'>
        {formatDate(row.original.createdAtUnix)}
      </div>
    ),
    enableSorting: false,
  },
  {
    id: 'actions',
    header: () => <span className='sr-only'>Actions</span>,
    cell: ({ row }) => <UsersRowActions user={row.original} />,
    enableSorting: false,
    enableHiding: false,
  },
]

function UsersRowActions({ user }: { user: User }) {
  const { setOpen, setCurrentRow } = useUsers()
  const currentUser = useAuthStore((s) => s.auth.user)
  const currentId = currentUser?.accountNo ?? ''
  const isSelf = String(user.id) === currentId
  const isCurrentUserOwner = Boolean(currentUser?.role?.includes('owner'))

  // Hierarchy rules for row actions:
  // - Edit & Reset Password:
  //   - Self: ALWAYS allowed
  //   - Target is owner (not self): DISABLED (Owner is strictly protected)
  //   - Target is admin (not self): ONLY Owner can edit/reset (Admin cannot edit another admin)
  //   - Target is agent/teknisi: Owner and Admin can edit/reset
  const canEdit = isSelf || (user.role !== 'owner' && (isCurrentUserOwner || user.role !== 'admin'))
  const canReset = isSelf || (user.role !== 'owner' && (isCurrentUserOwner || user.role !== 'admin'))

  // - Deactivate & Delete:
  //   - Self: ALWAYS disabled (prevent self lockout)
  //   - Target is owner: ALWAYS disabled (Owner account cannot be deleted/deactivated)
  //   - Target is admin: ONLY Owner can deactivate/delete
  //   - Target is agent/teknisi: Owner and Admin can deactivate/delete
  const canToggleOrDelete = !isSelf && user.role !== 'owner' && (isCurrentUserOwner || user.role !== 'admin')

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant='ghost' className='h-8 w-8 p-0' aria-label='Open menu'>
          <MoreHorizontal className='h-4 w-4' />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end' className='w-44'>
        <DropdownMenuLabel className='text-xs text-muted-foreground'>
          {user.username} {isSelf && '(You)'}
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          onClick={() => {
            setCurrentRow(user)
            setOpen('edit')
          }}
          disabled={!canEdit}
        >
          Edit user
        </DropdownMenuItem>
        <DropdownMenuItem
          onClick={() => {
            setCurrentRow(user)
            setOpen('reset')
          }}
          disabled={!canReset}
        >
          Reset password
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          onClick={() => {
            setCurrentRow(user)
            setOpen('toggle')
          }}
          disabled={!canToggleOrDelete}
        >
          {user.isActive ? (
            <>
              <ShieldOff className='me-2 h-4 w-4' /> Deactivate
            </>
          ) : (
            <>
              <ShieldCheck className='me-2 h-4 w-4' /> Activate
            </>
          )}
        </DropdownMenuItem>
        <DropdownMenuItem
          onClick={() => {
            setCurrentRow(user)
            setOpen('delete')
          }}
          disabled={!canToggleOrDelete}
          className='text-destructive focus:text-destructive'
        >
          Delete user
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
