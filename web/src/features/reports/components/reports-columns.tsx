import { type ColumnDef } from '@tanstack/react-table'
import { DataTableColumnHeader } from '@/components/data-table'
import { Badge } from '@/components/ui/badge'
import type { HotspotReport } from '@/gen/v1/hotspot_pb'
import { ReportsRowActions } from './reports-row-actions'

export const reportsColumns: ColumnDef<HotspotReport>[] = [
  {
    accessorKey: 'date',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Date & Time' />
    ),
    cell: ({ row }) => (
      <div className='flex flex-col text-xs'>
        <span className='font-medium'>{row.original.date}</span>
        <span className='text-[11px] text-muted-foreground font-mono'>
          {row.original.time}
        </span>
      </div>
    ),
    enableSorting: true,
  },
  {
    accessorKey: 'username',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Voucher / User' />
    ),
    cell: ({ row }) => (
      <span className='font-mono font-semibold'>{row.original.username}</span>
    ),
    enableSorting: true,
  },
  {
    accessorKey: 'profile',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Profile' />
    ),
    cell: ({ row }) => (
      <Badge variant='secondary' className='text-xs'>
        {row.original.profile}
      </Badge>
    ),
  },
  {
    accessorKey: 'price',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Price (Income)' />
    ),
    cell: ({ row }) => {
      const price = Number(row.original.price || 0)
      return (
        <span className='font-mono text-xs font-semibold text-emerald-600 dark:text-emerald-400'>
          Rp {price.toLocaleString('id-ID')}
        </span>
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
    cell: ({ row }) => <ReportsRowActions report={row.original} />,
  },
]
