import { type ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/ui/badge'
import { DataTableColumnHeader } from '@/components/data-table'
import type { HotspotProfile } from '@/gen/v1/hotspot_pb'
import { ProfilesRowActions } from './profiles-row-actions'

export const profilesColumns: ColumnDef<HotspotProfile>[] = [
  {
    accessorKey: 'name',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Profile Name' />
    ),
    cell: ({ row }) => <span className='font-semibold'>{row.original.name}</span>,
    enableSorting: true,
  },
  {
    accessorKey: 'sharedUsers',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Shared Users' />
    ),
    cell: ({ row }) => (
      <span className='font-mono text-xs'>{row.original.sharedUsers || '1'}</span>
    ),
  },
  {
    accessorKey: 'rateLimit',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Rate Limit' />
    ),
    cell: ({ row }) => (
      <span className='font-mono text-xs text-muted-foreground'>
        {row.original.rateLimit || '-'}
      </span>
    ),
  },
  {
    accessorKey: 'modeExpire',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Expire Mode' />
    ),
    cell: ({ row }) => {
      const mode = row.original.modeExpire
      if (!mode || mode === '0' || mode === 'None') {
        return <Badge variant='outline'>None</Badge>
      }
      if (mode.includes('rem')) {
        return <Badge variant='destructive' className='text-xs'>Remove</Badge>
      }
      return <Badge variant='secondary' className='text-xs'>Notice</Badge>
    },
  },
  {
    accessorKey: 'validity',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Validity' />
    ),
    cell: ({ row }) => (
      <span className='font-medium text-xs'>{row.original.validity || '-'}</span>
    ),
  },
  {
    accessorKey: 'price',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Price' />
    ),
    cell: ({ row }) => {
      const price = Number(row.original.price || 0)
      return (
        <span className='font-mono text-xs'>
          {price > 0 ? price.toLocaleString('id-ID') : '-'}
        </span>
      )
    },
  },
  {
    accessorKey: 'sellingPrice',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Selling Price' />
    ),
    cell: ({ row }) => {
      const sprice = Number(row.original.sellingPrice || 0)
      return (
        <span className='font-mono text-xs font-semibold text-emerald-600 dark:text-emerald-400'>
          {sprice > 0 ? sprice.toLocaleString('id-ID') : '-'}
        </span>
      )
    },
  },
  {
    accessorKey: 'lockUser',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Lock' />
    ),
    cell: ({ row }) => {
      const lockUser = row.original.lockUser === 'Enable'
      const lockServer = row.original.lockServer === 'Enable'
      if (!lockUser && !lockServer) return <span className='text-xs text-muted-foreground'>-</span>
      return (
        <div className='flex gap-1'>
          {lockUser && <Badge variant='outline' className='text-[10px]'>User</Badge>}
          {lockServer && <Badge variant='outline' className='text-[10px]'>Server</Badge>}
        </div>
      )
    },
  },
  {
    id: 'actions',
    cell: ({ row }) => <ProfilesRowActions profile={row.original} />,
  },
]
