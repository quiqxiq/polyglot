import { Badge } from '@/components/ui/badge'
import { DataTableColumnHeader } from '@/components/data-table'
import type { PPPSecret } from '@/gen/v1/ppp_pb'
import type { ColumnDef } from '@tanstack/react-table'
import { Server, Shield, UserX } from 'lucide-react'
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
  },
  {
    accessorKey: 'service',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title="Service" />
    ),
    cell: ({ row }) => {
      const service = row.getValue('service') as string
      return (
        <Badge variant="secondary" className="uppercase font-mono text-[10px]">
          <Server className="mr-1 h-3 w-3" />
          {service || 'any'}
        </Badge>
      )
    },
    filterFn: (row, id, value: string[]) => {
      const service = ((row.getValue(id) as string) || 'any').toLowerCase()
      return value.includes(service)
    },
  },
  {
    id: 'address',
    header: 'Assigned IP',
    cell: ({ row }) => {
      const remote = row.original.remoteAddress
      return (
        <span className="text-xs font-mono text-muted-foreground">
          {remote || 'Profile default'}
        </span>
      )
    },
  },
  {
    accessorKey: 'comment',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title="Comment" />
    ),
    cell: ({ row }) => {
      const comment = row.getValue('comment') as string
      if (!comment) return <span className="text-xs text-muted-foreground">-</span>
      return (
        <span className="max-w-[180px] truncate text-xs text-muted-foreground">
          {comment}
        </span>
      )
    },
    filterFn: (row, id, value: string[]) => {
      const comment = (row.getValue(id) as string) || ''
      return value.includes(comment)
    },
  },
  {
    accessorKey: 'disabled',
    id: 'status',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title="Status" />
    ),
    cell: ({ row }) => {
      const disabled = row.original.disabled
      return disabled ? (
        <Badge variant="destructive" className="text-[10px]">
          Disabled
        </Badge>
      ) : (
        <Badge variant="secondary" className="text-[10px] text-muted-foreground">
          Offline
        </Badge>
      )
    },
    filterFn: (row, _id, value: string[]) => {
      const disabled = row.original.disabled
      const status = disabled ? 'disabled' : 'offline'
      return value.includes(status)
    },
  },
  {
    accessorKey: 'lastLoggedOut',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title="Last Logout" />
    ),
    cell: ({ row }) => {
      const last = row.getValue('lastLoggedOut') as string
      return (
        <span className="text-xs text-muted-foreground font-mono">
          {last || 'Never'}
        </span>
      )
    },
  },
  {
    id: 'actions',
    cell: ({ row }) => <InactiveRowActions row={row.original} />,
  },
]
