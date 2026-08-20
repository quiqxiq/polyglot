import { type ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/ui/badge'
import { DataTableColumnHeader } from '@/components/data-table'
import type { HotspotHost } from '@/gen/v1/hotspot_pb'
import { HostsRowActions } from './hosts-row-actions'

export const hostsColumns: ColumnDef<HotspotHost>[] = [
  {
    id: 'flags',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Flag' />
    ),
    cell: ({ row }) => {
      const bypassed = row.original.bypassed
      const authorized = row.original.authorized
      return (
        <div className='flex gap-1'>
          {authorized && (
            <Badge className='bg-emerald-500 hover:bg-emerald-600 text-[10px] px-1 py-0' title='Authorized'>
              A
            </Badge>
          )}
          {bypassed && (
            <Badge variant='secondary' className='text-[10px] px-1 py-0 bg-sky-500/20 text-sky-600 dark:text-sky-400' title='Bypassed'>
              P
            </Badge>
          )}
          {!authorized && !bypassed && (
            <span className='text-xs text-muted-foreground'>-</span>
          )}
        </div>
      )
    },
  },
  {
    accessorKey: 'macAddress',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='MAC Address' />
    ),
    cell: ({ row }) => (
      <span className='font-mono text-xs font-semibold'>{row.original.macAddress}</span>
    ),
    enableSorting: true,
  },
  {
    accessorKey: 'address',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='IP Address' />
    ),
    cell: ({ row }) => (
      <span className='font-mono text-xs'>{row.original.address}</span>
    ),
  },
  {
    accessorKey: 'toAddress',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='To Address (NAT)' />
    ),
    cell: ({ row }) => (
      <span className='font-mono text-xs text-muted-foreground'>
        {row.original.toAddress || '-'}
      </span>
    ),
  },
  {
    accessorKey: 'server',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Server' />
    ),
    cell: ({ row }) => (
      <span className='text-xs text-muted-foreground'>
        {row.original.server || 'all'}
      </span>
    ),
  },
  {
    accessorKey: 'comment',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Comment' />
    ),
    cell: ({ row }) => (
      <span className='font-mono text-xs text-muted-foreground truncate max-w-[150px] block' title={row.original.comment}>
        {row.original.comment || '-'}
      </span>
    ),
  },
  {
    id: 'actions',
    cell: ({ row }) => <HostsRowActions host={row.original} />,
  },
]
