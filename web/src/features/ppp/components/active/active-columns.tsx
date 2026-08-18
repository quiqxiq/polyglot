import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
import { DataTableColumnHeader } from '@/components/data-table'
import type { PPPActiveSession } from '@/gen/v1/ppp_pb'
import type { ColumnDef } from '@tanstack/react-table'
import { Activity, Clock, Globe, Network, Server } from 'lucide-react'
import { ActiveRowActions } from './active-row-actions'

export const activeColumns: ColumnDef<PPPActiveSession>[] = [
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
    accessorKey: 'service',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title="Service" />
    ),
    cell: ({ row }) => {
      const service = row.getValue('service') as string
      return (
        <Badge variant="secondary" className="uppercase font-mono text-[10px]">
          <Server className="mr-1 h-3 w-3" />
          {service || 'pppoe'}
        </Badge>
      )
    },
    filterFn: (row, id, value: string[]) => {
      const service = ((row.getValue(id) as string) || 'pppoe').toLowerCase()
      return value.includes(service)
    },
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
  },
  {
    accessorKey: 'callerId',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title="Caller ID (MAC)" />
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
  },
  {
    id: 'traffic',
    header: 'Traffic In / Out',
    cell: ({ row }) => {
      const inBytes = row.original.limitBytesIn
      const outBytes = row.original.limitBytesOut
      if (!inBytes && !outBytes) {
        return <span className="text-xs text-muted-foreground font-mono">Unlimited</span>
      }
      return (
        <div className="flex flex-col text-[11px] font-mono text-muted-foreground">
          {inBytes && <span>↓ {inBytes}</span>}
          {outBytes && <span>↑ {outBytes}</span>}
        </div>
      )
    },
  },
  {
    accessorKey: 'radius',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title="Auth" />
    ),
    cell: ({ row }) => {
      const radius = row.getValue('radius') as boolean
      return radius ? (
        <Badge variant="outline" className="text-[10px] border-sky-500/30 bg-sky-500/10 text-sky-600">
          RADIUS
        </Badge>
      ) : (
        <Badge variant="outline" className="text-[10px] text-muted-foreground">
          Local
        </Badge>
      )
    },
    filterFn: (row, id, value: string[]) => {
      const radius = row.getValue(id) as boolean
      const authType = radius ? 'radius' : 'local'
      return value.includes(authType)
    },
  },
  {
    id: 'actions',
    cell: ({ row }) => <ActiveRowActions row={row.original} />,
  },
]
