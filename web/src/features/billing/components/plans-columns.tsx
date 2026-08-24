import { type ColumnDef } from '@tanstack/react-table'
import { Zap } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { DataTableColumnHeader } from '@/components/data-table'
import type { Plan } from '@/gen/v1/billing_pb'
import { PlansRowActions } from './plans-row-actions'

const SERVICE_TYPE_BADGE: Record<string, string> = {
  PPPOE: 'bg-blue-500/15 text-blue-700 dark:text-blue-400 border-blue-500/30',
  HOTSPOT:
    'bg-purple-500/15 text-purple-700 dark:text-purple-400 border-purple-500/30',
  DEDICATED:
    'bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border-emerald-500/30',
}

function fmtRate(kbps: number) {
  return kbps >= 1000 ? `${Math.round(kbps / 1000)}M` : `${kbps}k`
}

const idrFormat = new Intl.NumberFormat('id-ID', {
  style: 'currency',
  currency: 'IDR',
  maximumFractionDigits: 0,
})

export const plansColumns: ColumnDef<Plan>[] = [
  {
    accessorKey: 'name',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Nama' />
    ),
    cell: ({ row }) => (
      <span className='font-semibold'>{row.original.name}</span>
    ),
  },
  {
    accessorKey: 'serviceType',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Tipe' />
    ),
    cell: ({ row }) => (
      <Badge
        variant='outline'
        className={`text-xs ${SERVICE_TYPE_BADGE[row.original.serviceType] ?? ''}`}
      >
        {row.original.serviceType}
      </Badge>
    ),
    enableSorting: true,
  },
  {
    id: 'bandwidth',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Bandwidth' />
    ),
    cell: ({ row }) => (
      <span className='font-mono text-xs'>
        {fmtRate(row.original.bandwidthDownloadKbps)} /{' '}
        {fmtRate(row.original.bandwidthUploadKbps)}
      </span>
    ),
    enableSorting: false,
  },
  {
    accessorKey: 'price',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Harga' />
    ),
    cell: ({ row }) => (
      <span className='font-mono text-xs'>
        {idrFormat.format(row.original.price)}
      </span>
    ),
    enableSorting: true,
  },
  {
    accessorKey: 'validity',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Validity' />
    ),
    cell: ({ row }) => (
      <span className='text-xs text-muted-foreground'>
        {row.original.validity || '-'}
      </span>
    ),
    meta: { className: 'hidden md:table-cell' },
  },
  {
    id: 'burst',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Burst' />
    ),
    cell: ({ row }) =>
      row.original.burstDownloadKbps > 0 ? (
        <Badge variant='outline' className='gap-1 text-xs'>
          <Zap className='h-3 w-3' />
          burst
        </Badge>
      ) : (
        <span className='text-muted-foreground'>-</span>
      ),
    enableSorting: false,
  },
  {
    accessorKey: 'isActive',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Status' />
    ),
    cell: ({ row }) =>
      row.original.isActive ? (
        <Badge
          variant='outline'
          className='bg-emerald-500/15 text-xs text-emerald-700 dark:text-emerald-400 border-emerald-500/30'
        >
          Active
        </Badge>
      ) : (
        <Badge
          variant='outline'
          className='bg-slate-500/15 text-xs text-slate-700 dark:text-slate-400 border-slate-500/30'
        >
          Inactive
        </Badge>
      ),
  },
  {
    id: 'actions',
    cell: ({ row }) => <PlansRowActions plan={row.original} />,
    enableSorting: false,
  },
]
