import { useMemo } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/ui/badge'
import { DataTableColumnHeader } from '@/components/data-table'
import type { Registration } from '@/gen/v1/registration_pb'
import { usePlansQuery } from '@/features/billing/api/use-plans'
import { registrationStatusBadge } from '../data/constants'
import { RegistrationRowActions } from './registration-row-actions'
import { useRegistration } from './registration-provider'

function formatUnixDate(unix: bigint | number | undefined) {
  const num = Number(unix || 0)
  if (!num) return '-'
  return new Date(num * 1000).toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}

function RegistrationNameCell({ registration }: { registration: Registration }) {
  const { setOpen, setCurrentRow } = useRegistration()
  return (
    <button
      type='button'
      onClick={() => {
        setCurrentRow(registration)
        setOpen('detail')
      }}
      className='text-left font-semibold hover:underline hover:text-primary transition-colors'
    >
      {registration.fullName}
    </button>
  )
}

export function useRegistrationColumns(): ColumnDef<Registration>[] {
  const { data: plans = [] } = usePlansQuery(false)

  const planNameById = useMemo(() => {
    const map = new Map<string, string>()
    for (const plan of plans) map.set(plan.id, plan.name)
    return map
  }, [plans])

  return useMemo(
    () => [
      {
        accessorKey: 'registrationNo',
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title='No. Reg' />
        ),
        cell: ({ row }) => (
          <span className='font-mono text-xs font-semibold'>
            {row.original.registrationNo}
          </span>
        ),
        enableSorting: true,
      },
      {
        accessorKey: 'fullName',
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title='Calon Pelanggan' />
        ),
        cell: ({ row }) => <RegistrationNameCell registration={row.original} />,
      },
      {
        accessorKey: 'phone',
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title='WhatsApp' />
        ),
        cell: ({ row }) => (
          <span className='font-mono text-xs'>{row.original.phone || '-'}</span>
        ),
      },
      {
        accessorKey: 'planId',
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title='Paket Diminta' />
        ),
        cell: ({ row }) => (
          <span className='text-sm font-medium'>
            {planNameById.get(row.original.planId) || row.original.planId || '-'}
          </span>
        ),
      },
      {
        accessorKey: 'address',
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title='Alamat Pemasangan' />
        ),
        cell: ({ row }) => (
          <span className='line-clamp-1 max-w-xs text-xs text-muted-foreground'>
            {row.original.address || '-'}
          </span>
        ),
        meta: { className: 'hidden md:table-cell' },
      },
      {
        accessorKey: 'status',
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title='Status Pipeline' />
        ),
        cell: ({ row }) => {
          const badge = registrationStatusBadge(row.original.status)
          return (
            <Badge variant='outline' className={`text-xs ${badge.className}`}>
              {badge.label}
            </Badge>
          )
        },
      },
      {
        accessorKey: 'scheduledInstallDateUnix',
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title='Jadwal Pasang' />
        ),
        cell: ({ row }) => {
          const dateStr = formatUnixDate(row.original.scheduledInstallDateUnix)
          const timeStr = row.original.scheduledInstallTime || ''
          if (dateStr === '-') return <span className='text-xs text-muted-foreground'>-</span>
          return (
            <div className='text-xs'>
              <span className='font-medium'>{dateStr}</span>
              {timeStr && <span className='text-muted-foreground'> ({timeStr})</span>}
            </div>
          )
        },
        meta: { className: 'hidden lg:table-cell' },
      },
      {
        accessorKey: 'targetDeviceId',
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title='Router Terpasang' />
        ),
        cell: ({ row }) => {
          const label = row.original.targetDeviceName || row.original.targetDeviceId || '-'
          return <span className='font-mono text-xs text-muted-foreground'>{label}</span>
        },
        meta: { className: 'hidden xl:table-cell' },
      },
      {
        id: 'actions',
        cell: ({ row }) => <RegistrationRowActions registration={row.original} />,
        enableSorting: false,
      },
    ],
    [planNameById]
  )
}
