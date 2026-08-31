import { useMemo, useState } from 'react'
import {
  type SortingState,
  type VisibilityState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from '@tanstack/react-table'
import { Cross2Icon } from '@radix-ui/react-icons'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { DataTablePagination } from '@/components/data-table'
import type { Invoice } from '@/gen/v1/billing_pb'
import type { Customer } from '@/gen/v1/customer_pb'
import { useInvoiceColumns } from './invoices-columns'
import { useInvoices } from './invoices-provider'

interface InvoicesTableProps {
  data: Invoice[]
  customers: Customer[]
  isLoading?: boolean
}

export function InvoicesTable({ data, customers, isLoading }: InvoicesTableProps) {
  const { filters, setFilters } = useInvoices()
  const [sorting, setSorting] = useState<SortingState>([])
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})
  const [globalFilter, setGlobalFilter] = useState('')

  const customerMap = useMemo(() => new Map(customers.map((c) => [c.id, c])), [customers])
  const columns = useInvoiceColumns(customerMap)

  // Filter berdasarkan customer search, nomor faktur, status, dan periode
  const filteredData = useMemo(() => {
    return data.filter((inv) => {
      if (filters.status && filters.status !== 'ALL' && inv.status !== filters.status) {
        return false
      }
      if (filters.period && !inv.period.includes(filters.period)) {
        return false
      }
      if (filters.customerId && inv.customerId !== filters.customerId) {
        return false
      }
      return true
    })
  }, [data, filters])

  const table = useReactTable({
    data: filteredData,
    columns,
    state: {
      sorting,
      columnVisibility,
      globalFilter,
    },
    onSortingChange: setSorting,
    onColumnVisibilityChange: setColumnVisibility,
    onGlobalFilterChange: setGlobalFilter,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
  })

  return (
    <div className='flex flex-1 flex-col gap-4'>
      {/* Search & Filter Toolbar */}
      <div className='flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between'>
        <div className='flex flex-1 flex-wrap items-center gap-2'>
          {/* Search Box */}
          <div className='relative w-full sm:w-72'>
            <Input
              placeholder='Cari nomor faktur, kode bayar, atau pelanggan...'
              className='h-9 pr-8 text-xs sm:text-sm'
              value={globalFilter}
              onChange={(e) => setGlobalFilter(e.target.value)}
            />
            {globalFilter && (
              <button
                type='button'
                onClick={() => setGlobalFilter('')}
                className='absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground'
                title='Clear search'
              >
                <Cross2Icon className='h-3.5 w-3.5' />
              </button>
            )}
          </div>

          {/* Filter Status Tagihan */}
          <Select
            value={filters.status || 'ALL'}
            onValueChange={(val) =>
              setFilters((prev) => ({ ...prev, status: val === 'ALL' ? '' : val }))
            }
          >
            <SelectTrigger className='h-9 w-full sm:w-44 text-xs'>
              <SelectValue placeholder='Semua Status' />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value='ALL'>Semua Status</SelectItem>
              <SelectItem value='UNPAID'>Belum Bayar (UNPAID)</SelectItem>
              <SelectItem value='PARTIAL'>Sebagian (PARTIAL)</SelectItem>
              <SelectItem value='PAID'>Lunas (PAID)</SelectItem>
              <SelectItem value='OVERDUE'>Jatuh Tempo (OVERDUE)</SelectItem>
              <SelectItem value='CANCELLED'>Dibatalkan (CANCELLED)</SelectItem>
            </SelectContent>
          </Select>

          {/* Filter Periode */}
          <Input
            placeholder='Periode (YYYY-MM)'
            className='h-9 w-full sm:w-40 text-xs font-mono'
            value={filters.period}
            onChange={(e) => setFilters((prev) => ({ ...prev, period: e.target.value }))}
          />
        </div>
      </div>

      {/* Table Container */}
      <div className='overflow-hidden rounded-md border bg-card'>
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead key={header.id}>
                    {header.isPlaceholder
                      ? null
                      : flexRender(header.column.columnDef.header, header.getContext())}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={columns.length} className='h-24 text-center text-muted-foreground'>
                  Memuat data faktur tagihan...
                </TableCell>
              </TableRow>
            ) : table.getRowModel().rows.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow key={row.id}>
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={columns.length} className='h-24 text-center text-muted-foreground'>
                  Belum ada faktur tagihan yang sesuai filter
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      <DataTablePagination table={table} className='mt-auto' />
    </div>
  )
}
