import { type ColumnDef } from '@tanstack/react-table'
import { DataTableColumnHeader } from '@/components/data-table'
import type { HotspotActiveSession } from '@/gen/v1/hotspot_pb'
import { ActiveRowActions } from './active-row-actions'

function formatBytes(bytesStr: string): string {
  const bytes = Number(bytesStr || 0)
  if (isNaN(bytes) || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`
}

export const activeColumns: ColumnDef<HotspotActiveSession>[] = [
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
    accessorKey: 'user',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='User' />
    ),
    cell: ({ row }) => (
      <span className='font-mono font-semibold'>{row.original.user}</span>
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
    accessorKey: 'macAddress',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='MAC Address' />
    ),
    cell: ({ row }) => (
      <span className='font-mono text-xs text-muted-foreground'>
        {row.original.macAddress || '-'}
      </span>
    ),
  },
  {
    accessorKey: 'uptime',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Uptime' />
    ),
    cell: ({ row }) => (
      <span className='font-mono text-xs font-medium'>{row.original.uptime}</span>
    ),
  },
  {
    accessorKey: 'bytesOut',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Traffic (In / Out)' />
    ),
    cell: ({ row }) => {
      return (
        <div className='flex flex-col text-xs font-mono'>
          <span className='text-emerald-600 dark:text-emerald-400'>
            ↓ {formatBytes(row.original.bytesOut)}
          </span>
          <span className='text-sky-600 dark:text-sky-400'>
            ↑ {formatBytes(row.original.bytesIn)}
          </span>
        </div>
      )
    },
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
    cell: ({ row }) => <ActiveRowActions session={row.original} />,
  },
]
