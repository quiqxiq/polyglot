import { type ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/ui/badge'
import { DataTableColumnHeader } from '@/components/data-table'
import type { HotspotUser } from '@/gen/v1/hotspot_pb'
import { InactiveRowActions } from './inactive-row-actions'

function formatBytes(bytesStr: string): string {
  const bytes = Number(bytesStr || 0)
  if (isNaN(bytes) || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`
}

export const inactiveColumns: ColumnDef<HotspotUser>[] = [
  {
    accessorKey: 'name',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Username / Code' />
    ),
    cell: ({ row }) => {
      const isVoucher = row.original.name === row.original.password
      return (
        <div className='flex items-center gap-2'>
          <span className='font-mono font-medium'>{row.original.name}</span>
          {isVoucher && (
            <Badge variant='outline' className='text-[10px] py-0 px-1 font-normal'>
              vc
            </Badge>
          )}
        </div>
      )
    },
    enableSorting: true,
  },
  {
    accessorKey: 'profile',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Profile' />
    ),
    cell: ({ row }) => (
      <Badge variant='secondary' className='font-normal text-xs'>
        {row.original.profile || 'default'}
      </Badge>
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
    accessorKey: 'limitUptime',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Time / Quota Limit' />
    ),
    cell: ({ row }) => {
      const limitTime = row.original.limitUptime
      const limitBytes = row.original.limitBytes
      return (
        <div className='flex flex-col text-xs'>
          <span>{limitTime ? `Time: ${limitTime}` : 'Unlimited time'}</span>
          {limitBytes && (
            <span className='text-[11px] text-muted-foreground font-mono'>
              Quota: {formatBytes(limitBytes)}
            </span>
          )}
        </div>
      )
    },
  },
  {
    accessorKey: 'uptime',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Used Uptime' />
    ),
    cell: ({ row }) => (
      <span className='font-mono text-xs text-muted-foreground'>
        {row.original.uptime || '0s'}
      </span>
    ),
  },
  {
    accessorKey: 'comment',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Comment / Batch Tag' />
    ),
    cell: ({ row }) => (
      <span className='font-mono text-xs text-muted-foreground truncate max-w-[180px] block' title={row.original.comment}>
        {row.original.comment || '-'}
      </span>
    ),
  },
  {
    id: 'actions',
    cell: ({ row }) => <InactiveRowActions user={row.original} />,
  },
]
