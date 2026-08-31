import { useMemo } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/ui/badge'
import { DataTableColumnHeader } from '@/components/data-table'
import type { Customer } from '@/gen/v1/customer_pb'
import { useSubscriptionsQuery, useInvoicesQuery } from '@/features/billing/api/use-billing'
import { customerStatusBadge } from '../data/constants'
import { CustomersRowActions } from './customers-row-actions'
import { useCustomers } from './customers-provider'

function formatRegisteredAt(value: Customer['registeredAtUnix']) {
  const unix = Number(value || 0)
  if (!unix) return '-'
  return new Date(unix * 1000).toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}

function CustomerNameCell({ customer }: { customer: Customer }) {
  const { setOpen, setCurrentRow } = useCustomers()
  return (
    <button
      type='button'
      onClick={() => {
        setCurrentRow(customer)
        setOpen('detail')
      }}
      className='text-left font-semibold hover:underline hover:text-primary transition-colors'
    >
      {customer.name}
    </button>
  )
}

export function useCustomersColumns(): ColumnDef<Customer>[] {
  const { data: allSubscriptions = [] } = useSubscriptionsQuery('', { enabled: true })
  const { data: allInvoices = [] } = useInvoicesQuery('', '', { enabled: true })

  const subCountByCustomer = useMemo(() => {
    const map = new Map<string, number>()
    for (const s of allSubscriptions) {
      if (s.status !== 'TERMINATED' && s.status !== 'CANCELLED') {
        map.set(s.customerId, (map.get(s.customerId) || 0) + 1)
      }
    }
    return map
  }, [allSubscriptions])

  const unpaidCountByCustomer = useMemo(() => {
    const map = new Map<string, number>()
    for (const inv of allInvoices) {
      if (inv.status !== 'PAID') {
        map.set(inv.customerId, (map.get(inv.customerId) || 0) + 1)
      }
    }
    return map
  }, [allInvoices])

  return useMemo(
    () => [
      {
        accessorKey: 'customerCode',
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title='Kode' />
        ),
        cell: ({ row }) => (
          <span className='font-mono text-xs'>{row.original.customerCode}</span>
        ),
        enableSorting: true,
      },
      {
        accessorKey: 'name',
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title='Nama' />
        ),
        cell: ({ row }) => <CustomerNameCell customer={row.original} />,
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
        accessorKey: 'activeSubscriptionsCount',
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title='Langganan' />
        ),
        cell: ({ row }) => {
          const liveCount = subCountByCustomer.get(row.original.id)
          const count = liveCount !== undefined ? liveCount : (row.original.activeSubscriptionsCount || 0)
          return (
            <Badge variant={count > 0 ? 'secondary' : 'outline'} className='text-xs'>
              {count} Layanan
            </Badge>
          )
        },
      },
      {
        accessorKey: 'unpaidInvoicesCount',
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title='Tagihan' />
        ),
        cell: ({ row }) => {
          const liveCount = unpaidCountByCustomer.get(row.original.id)
          const count = liveCount !== undefined ? liveCount : (row.original.unpaidInvoicesCount || 0)
          if (count > 0) {
            return (
              <Badge
                variant='outline'
                className='border-amber-500/30 bg-amber-500/15 text-xs text-amber-700 dark:text-amber-400'
              >
                {count} Belum Bayar
              </Badge>
            )
          }
          return <span className='text-xs text-muted-foreground'>Tidak Ada</span>
        },
        meta: { className: 'hidden sm:table-cell' },
      },
      {
        accessorKey: 'status',
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title='Status' />
        ),
        cell: ({ row }) => {
          const badge = customerStatusBadge(row.original.status)
          return (
            <Badge variant='outline' className={`text-xs ${badge.className}`}>
              {badge.label}
            </Badge>
          )
        },
      },
      {
        accessorKey: 'portalAccessCode',
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title='Portal Access' />
        ),
        cell: ({ row }) => (
          <span className='font-mono text-xs'>
            {row.original.portalAccessCode || '-'}
          </span>
        ),
        meta: { className: 'hidden lg:table-cell' },
      },
      {
        accessorKey: 'registeredAtUnix',
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title='Terdaftar' />
        ),
        cell: ({ row }) => (
          <span className='text-xs text-muted-foreground'>
            {formatRegisteredAt(row.original.registeredAtUnix || row.original.createdAtUnix)}
          </span>
        ),
        meta: { className: 'hidden md:table-cell' },
      },
      {
        id: 'actions',
        cell: ({ row }) => <CustomersRowActions customer={row.original} />,
        enableSorting: false,
      },
    ],
    [subCountByCustomer, unpaidCountByCustomer]
  )
}

export const customersColumns: ColumnDef<Customer>[] = [
  {
    accessorKey: 'customerCode',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Kode' />
    ),
    cell: ({ row }) => (
      <span className='font-mono text-xs'>{row.original.customerCode}</span>
    ),
    enableSorting: true,
  },
  {
    accessorKey: 'name',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Nama' />
    ),
    cell: ({ row }) => <CustomerNameCell customer={row.original} />,
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
    accessorKey: 'activeSubscriptionsCount',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Langganan' />
    ),
    cell: ({ row }) => {
      const count = row.original.activeSubscriptionsCount || 0
      return (
        <Badge variant={count > 0 ? 'secondary' : 'outline'} className='text-xs'>
          {count} Layanan
        </Badge>
      )
    },
  },
  {
    accessorKey: 'unpaidInvoicesCount',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Tagihan' />
    ),
    cell: ({ row }) => {
      const count = row.original.unpaidInvoicesCount || 0
      if (count > 0) {
        return (
          <Badge
            variant='outline'
            className='border-amber-500/30 bg-amber-500/15 text-xs text-amber-700 dark:text-amber-400'
          >
            {count} Belum Bayar
          </Badge>
        )
      }
      return <span className='text-xs text-muted-foreground'>Tidak Ada</span>
    },
    meta: { className: 'hidden sm:table-cell' },
  },
  {
    accessorKey: 'status',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Status' />
    ),
    cell: ({ row }) => {
      const badge = customerStatusBadge(row.original.status)
      return (
        <Badge variant='outline' className={`text-xs ${badge.className}`}>
          {badge.label}
        </Badge>
      )
    },
  },
  {
    accessorKey: 'portalAccessCode',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Portal Access' />
    ),
    cell: ({ row }) => (
      <span className='font-mono text-xs'>
        {row.original.portalAccessCode || '-'}
      </span>
    ),
    meta: { className: 'hidden lg:table-cell' },
  },
  {
    accessorKey: 'registeredAtUnix',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Terdaftar' />
    ),
    cell: ({ row }) => (
      <span className='text-xs text-muted-foreground'>
        {formatRegisteredAt(row.original.registeredAtUnix || row.original.createdAtUnix)}
      </span>
    ),
    meta: { className: 'hidden md:table-cell' },
  },
  {
    id: 'actions',
    cell: ({ row }) => <CustomersRowActions customer={row.original} />,
    enableSorting: false,
  },
]
