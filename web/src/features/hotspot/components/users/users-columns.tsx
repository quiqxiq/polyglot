import { type ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
import { DataTableColumnHeader } from '@/components/data-table'
import type { HotspotUser } from '@/gen/v1/hotspot_pb'
import { UsersRowActions } from './users-row-actions'

function formatBytes(bytesStr: string): string {
  const bytes = Number(bytesStr || 0)
  if (isNaN(bytes) || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`
}

export const usersColumns: ColumnDef<HotspotUser>[] = [
  {
    id: 'select',
    header: ({ table }) => (
      <Checkbox
        checked={
          table.getIsAllPageRowsSelected() ||
          (table.getIsSomePageRowsSelected() && 'indeterminate')
        }
        onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
        aria-label='Select all'
        className='translate-y-0.5'
      />
    ),
    cell: ({ row }) => (
      <Checkbox
        checked={row.getIsSelected()}
        onCheckedChange={(value) => row.toggleSelected(!!value)}
        aria-label='Select row'
        className='translate-y-0.5'
      />
    ),
    enableSorting: false,
    enableHiding: false,
  },
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
    filterFn: (row, id, value) => {
      const profile = (row.getValue(id) as string) || 'default'
      if (Array.isArray(value)) {
        return value.includes(profile)
      }
      return profile === value
    },
    enableSorting: true,
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
    enableSorting: true,
  },
  {
    accessorKey: 'uptime',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Uptime / Limit' />
    ),
    cell: ({ row }) => {
      const uptime = row.original.uptime || '0s'
      const limit = row.original.limitUptime
      return (
        <div className='flex flex-col text-xs'>
          <span className='font-mono'>{uptime}</span>
          {limit && (
            <span className='text-[11px] text-muted-foreground'>
              limit: {limit}
            </span>
          )}
        </div>
      )
    },
    enableSorting: true,
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
    enableSorting: true,
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
    filterFn: (row, id, value) => {
      const comment = (row.getValue(id) as string) || ''
      if (Array.isArray(value)) {
        return value.some((v) => comment === v || comment.includes(v))
      }
      return comment === value || comment.includes(value)
    },
    enableSorting: true,
  },
  {
    id: 'actions',
    cell: ({ row }) => <UsersRowActions user={row.original} />,
  },
]
