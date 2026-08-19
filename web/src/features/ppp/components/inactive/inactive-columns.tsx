import { Badge } from '@/components/ui/badge'
import { DataTableColumnHeader } from '@/components/data-table'
import type { PPPSecret } from '@/gen/v1/ppp_pb'
import type { ColumnDef } from '@tanstack/react-table'
import { Network, Shield, UserX } from 'lucide-react'
import { InactiveRowActions } from './inactive-row-actions'

export const inactiveColumns: ColumnDef<PPPSecret>[] = [
  {
    accessorKey: 'name',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title="Username" />
    ),
    cell: ({ row }) => {
      const name = row.getValue('name') as string
      return (
        <div className="flex items-center gap-2 font-medium">
          <div className="flex h-7 w-7 items-center justify-center rounded bg-muted text-muted-foreground">
            <UserX className="h-3.5 w-3.5" />
          </div>
          <span className="font-mono text-sm">{name}</span>
        </div>
      )
    },
    enableSorting: true,
  },
  {
    accessorKey: 'lastLoggedOut',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title="Last Logged Out" />
    ),
    cell: ({ row }) => {
      const last = row.getValue('lastLoggedOut') as string
      return (
        <span className="text-xs text-muted-foreground font-mono">
          {last || 'Never'}
        </span>
      )
    },
    enableSorting: true,
  },
  {
    accessorKey: 'callerId',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title="MAC Address" />
    ),
    cell: ({ row }) => {
      const callerId = row.getValue('callerId') as string
      return (
        <div className="flex items-center gap-1.5 font-mono text-xs text-muted-foreground">
          <Network className="h-3.5 w-3.5" />
          <span>{callerId || '-'}</span>
        </div>
      )
    },
    enableSorting: true,
  },
  {
    accessorKey: 'profile',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title="Profile" />
    ),
    cell: ({ row }) => {
      const profile = row.getValue('profile') as string
      return (
        <Badge variant="outline" className="font-mono text-xs">
          <Shield className="mr-1 h-3 w-3 text-muted-foreground" />
          {profile || 'default'}
        </Badge>
      )
    },
    filterFn: (row, id, value: string[]) => {
      const profile = (row.getValue(id) as string) || 'default'
      return value.includes(profile)
    },
    enableSorting: true,
  },
  {
    id: 'actions',
    header: () => <div className="text-right pr-2">Actions</div>,
    cell: ({ row }) => <InactiveRowActions row={row.original} />,
  },
]
