import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
import { DataTableColumnHeader } from '@/components/data-table'
import type { PPPSecret } from '@/gen/v1/ppp_pb'
import type { ColumnDef } from '@tanstack/react-table'
import { Server, User } from 'lucide-react'
import { SecretsRowActions } from './secrets-row-actions'

export const secretsColumns: ColumnDef<PPPSecret>[] = [
  {
    id: 'select',
    header: ({ table }) => (
      <Checkbox
        checked={
          table.getIsAllPageRowsSelected() ||
          (table.getIsSomePageRowsSelected() && 'indeterminate')
        }
        onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
        aria-label="Select all"
        className="translate-y-[2px]"
      />
    ),
    cell: ({ row }) => (
      <Checkbox
        checked={row.getIsSelected()}
        onCheckedChange={(value) => row.toggleSelected(!!value)}
        aria-label="Select row"
        className="translate-y-[2px]"
      />
    ),
    enableSorting: false,
    enableHiding: false,
  },
  {
    accessorKey: 'name',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title="Username" />
    ),
    cell: ({ row }) => {
      const name = row.getValue('name') as string
      return (
        <div className="flex items-center gap-2 font-medium">
          <div className="flex h-7 w-7 items-center justify-center rounded bg-primary/10 text-primary">
            <User className="h-3.5 w-3.5" />
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
    header: 'IP / Remote Address',
    cell: ({ row }) => {
      const local = row.original.localAddress
      const remote = row.original.remoteAddress
      if (!local && !remote) {
        return <span className="text-xs text-muted-foreground">Profile default</span>
      }
      return (
        <div className="flex flex-col text-xs font-mono">
          {remote && <span>Remote: {remote}</span>}
          {local && <span className="text-muted-foreground">Local: {local}</span>}
        </div>
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
        <span className="max-w-[160px] truncate text-xs text-muted-foreground">
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
    accessorKey: 'lastLoggedOut',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title="Last Logout" />
    ),
    cell: ({ row }) => {
      const last = row.getValue('lastLoggedOut') as string
      return (
        <span className="text-xs text-muted-foreground">
          {last || 'Never'}
        </span>
      )
    },
  },
  {
    id: 'actions',
    cell: ({ row }) => <SecretsRowActions row={row.original} />,
  },
]
