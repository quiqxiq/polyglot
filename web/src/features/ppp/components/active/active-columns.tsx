import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
import { DataTableColumnHeader } from '@/components/data-table'
import type { EnrichedPPPActiveSession } from '../../api/use-ppp-stream'
import type { ColumnDef } from '@tanstack/react-table'
import { Activity, Clock, Globe, Network } from 'lucide-react'
import { ActiveRowActions } from './active-row-actions'

export const activeColumns: ColumnDef<EnrichedPPPActiveSession>[] = [
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
          <div className="flex h-7 w-7 items-center justify-center rounded bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
            <Activity className="h-3.5 w-3.5 animate-pulse" />
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
      const profile = row.original.profile || 'default'
      return (
        <Badge variant="outline" className="font-mono text-xs">
          {profile}
        </Badge>
      )
    },
    filterFn: (row, _id, value: string[]) => {
      const profile = row.original.profile || 'default'
      return value.includes(profile)
    },
    enableSorting: true,
  },
  {
    accessorKey: 'address',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title="IP Address" />
    ),
    cell: ({ row }) => {
      const address = row.getValue('address') as string
      return (
        <div className="flex items-center gap-1.5 font-mono text-xs">
          <Globe className="h-3.5 w-3.5 text-muted-foreground" />
          <span>{address || '-'}</span>
        </div>
      )
    },
    enableSorting: true,
  },
  {
    accessorKey: 'uptime',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title="Uptime" />
    ),
    cell: ({ row }) => {
      const uptime = row.getValue('uptime') as string
      return (
        <div className="flex items-center gap-1.5 text-xs text-muted-foreground font-mono">
          <Clock className="h-3.5 w-3.5 text-primary" />
          <span>{uptime || '-'}</span>
        </div>
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
    id: 'actions',
    header: () => <div className="text-right pr-2">Actions</div>,
    cell: ({ row }) => <ActiveRowActions row={row.original} />,
  },
]
