import { type ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/ui/badge'
import { DataTableColumnHeader } from '@/components/data-table'
import type { Customer } from '@/gen/v1/customer_pb'
import { customerStatusBadge } from '../data/constants'
import { CustomersRowActions } from './customers-row-actions'

function formatRegisteredAt(value: Customer['registeredAtUnix']) {
  const unix = Number(value || 0)
  if (!unix) return '-'
  return new Date(unix * 1000).toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
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
    cell: ({ row }) => (
      <span className='font-semibold'>{row.original.name}</span>
    ),
  },
  {
    accessorKey: 'phone',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Phone' />
    ),
    cell: ({ row }) => (
      <span className='font-mono text-xs'>{row.original.phone || '-'}</span>
    ),
  },
  {
    accessorKey: 'email',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Email' />
    ),
    cell: ({ row }) => (
      <span className='text-muted-foreground'>{row.original.email || '-'}</span>
    ),
    meta: { className: 'hidden md:table-cell' },
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
        {formatRegisteredAt(row.original.registeredAtUnix)}
      </span>
    ),
  },
  {
    id: 'actions',
    cell: ({ row }) => <CustomersRowActions customer={row.original} />,
    enableSorting: false,
  },
]
