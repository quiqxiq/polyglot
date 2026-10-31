import { type ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/ui/badge'
import { DataTableColumnHeader } from '@/components/data-table'
import { createSelectColumn } from '@/components/data-table/select-column'
import { Plan } from '@/gen/v1/plan_pb'
import { PlansRowActions } from './plans-row-actions'
import { Users, Zap } from 'lucide-react'

const SERVICE_TYPE_BADGE: Record<string, string> = {
  PPPOE: 'bg-blue-500/15 text-blue-700 dark:text-blue-400 border-blue-500/30',
  HOTSPOT: 'bg-orange-500/15 text-orange-700 dark:text-orange-400 border-orange-500/30',
  STATIC: 'bg-purple-500/15 text-purple-700 dark:text-purple-400 border-purple-500/30',
}

function fmtRate(kbps: number): string {
  if (kbps >= 1000) return `${(kbps / 1000).toFixed(0)} Mbps`
  return `${kbps} Kbps`
}

const idrFormat = new Intl.NumberFormat('id-ID', {
  style: 'currency',
  currency: 'IDR',
  maximumFractionDigits: 0,
})

export const getPlansColumns = (t: (key: string) => string): ColumnDef<Plan>[] => [
  createSelectColumn<Plan>(),
  {
    accessorKey: 'name',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title={t('billing.plan_name')} />
    ),
    cell: ({ row }) => (
      <span className='font-semibold'>{row.original.name}</span>
    ),
  },
  {
    accessorKey: 'serviceType',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title={t('common.type')} />
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
      <DataTableColumnHeader column={column} title='Bandwidth (DL / UL)' />
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
      <DataTableColumnHeader column={column} title={t('billing.price')} />
    ),
    cell: ({ row }) => (
      <span className='font-mono text-xs'>
        {idrFormat.format(row.original.price)}
      </span>
    ),
    enableSorting: true,
  },
  {
    accessorKey: 'sharedUsers',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title={t('common.details')} />
    ),
    cell: ({ row }) => {
      if (row.original.serviceType !== 'HOTSPOT') {
        return <span className='text-xs text-muted-foreground'>-</span>
      }
      return (
        <Badge variant='outline' className='gap-1 text-xs'>
          <Users className='h-3 w-3' />
          {row.original.sharedUsers || 1} Device
        </Badge>
      )
    },
    meta: { className: 'hidden md:table-cell' },
  },
  {
    id: 'burst',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title={t('billing.burst')} />
    ),
    cell: ({ row }) =>
      row.original.burstDownloadKbps > 0 ? (
        <Badge variant='outline' className='gap-1 text-xs'>
          <Zap className='h-3 w-3' />
          burst
        </Badge>
      ) : (
        <span className='text-muted-foreground text-xs'>-</span>
      ),
    enableSorting: false,
  },
  {
    accessorKey: 'isActive',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title={t('common.status')} />
    ),
    cell: ({ row }) =>
      row.original.isActive ? (
        <Badge
          variant='outline'
          className='bg-emerald-500/15 text-xs text-emerald-700 dark:text-emerald-400 border-emerald-500/30'
        >
          {t('common.active')}
        </Badge>
      ) : (
        <Badge
          variant='outline'
          className='bg-slate-500/15 text-xs text-slate-700 dark:text-slate-400 border-slate-500/30'
        >
          {t('common.inactive')}
        </Badge>
      ),
  },
  {
    id: 'actions',
    cell: ({ row }) => <PlansRowActions plan={row.original} />,
    enableSorting: false,
  },
]

export const plansColumns = getPlansColumns((k) => {
  const parts = k.split('.')
  return parts[parts.length - 1] || k
})
