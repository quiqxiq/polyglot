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
  const isSelf = String(user.id) === currentUserID()

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant='ghost' className='h-8 w-8 p-0' aria-label='Open menu'>
          <MoreHorizontal className='h-4 w-4' />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end' className='w-44'>
        <DropdownMenuLabel className='text-xs text-muted-foreground'>
          {user.username}
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          onClick={() => {
            setCurrentRow(user)
            setOpen('edit')
          }}
          disabled={isSelf}
        >
          Edit user
        </DropdownMenuItem>
        <DropdownMenuItem
          onClick={() => {
            setCurrentRow(user)
            setOpen('reset')
          }}
        >
          Reset password
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          onClick={() => {
            setCurrentRow(user)
            setOpen('toggle')
          }}
          disabled={isSelf}
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
          disabled={isSelf}
          className='text-destructive focus:text-destructive'
        >
          Delete user
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

// currentUserID membaca id user yang sedang login dari auth store (field
// accountNo dipakai untuk menyimpan id string — lihat user-auth-form).
function currentUserID(): string {
  const user = useAuthStore.getState().auth.user
  return user?.accountNo ?? ''
}
