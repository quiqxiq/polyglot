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
import type { CashTransaction, CashAccount, CashCategory } from '@/gen/v1/cashbook_pb'
import { createTransactionColumns } from './cashbook-transactions-columns'
import { useCashbook } from './cashbook-provider'

interface CashbookTransactionsTableProps {
  data: CashTransaction[]
  accounts: CashAccount[]
  categories: CashCategory[]
  isLoading?: boolean
}

export function CashbookTransactionsTable({
  data,
  accounts,
  categories,
  isLoading,
}: CashbookTransactionsTableProps) {
  const { filters, setFilters } = useCashbook()
  const [sorting, setSorting] = useState<SortingState>([])
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})
  const [globalFilter, setGlobalFilter] = useState('')

  const accountMap = useMemo(() => new Map(accounts.map((a) => [a.id, a])), [accounts])
  const categoryMap = useMemo(() => new Map(categories.map((c) => [c.id, c])), [categories])

  const columns = useMemo(
    () => createTransactionColumns(accountMap, categoryMap),
    [accountMap, categoryMap]
  )

  const table = useReactTable({
    data,
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
              placeholder='Cari nomor transaksi, deskripsi...'
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

          {/* Filter Rekening */}
          <Select
            value={filters.accountId || 'ALL'}
            onValueChange={(val) =>
              setFilters((prev) => ({ ...prev, accountId: val === 'ALL' ? '' : val }))
            }
          >
            <SelectTrigger className='h-9 w-full sm:w-48 text-xs'>
              <SelectValue placeholder='Semua Rekening' />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value='ALL'>Semua Rekening Kas</SelectItem>
              {accounts.map((acc) => (
                <SelectItem key={acc.id} value={acc.id}>
                  {acc.name} ({acc.accountCode})
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          {/* Filter Kategori */}
          <Select
            value={filters.categoryId || 'ALL'}
            onValueChange={(val) =>
              setFilters((prev) => ({ ...prev, categoryId: val === 'ALL' ? '' : val }))
            }
          >
            <SelectTrigger className='h-9 w-full sm:w-44 text-xs'>
              <SelectValue placeholder='Semua Kategori' />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value='ALL'>Semua Kategori</SelectItem>
              {categories.map((cat) => (
                <SelectItem key={cat.id} value={cat.id}>
                  {cat.name} ({cat.type})
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          {/* Filter Arah Transaksi */}
          <Select
            value={filters.direction || 'ALL'}
            onValueChange={(val) =>
              setFilters((prev) => ({ ...prev, direction: val === 'ALL' ? '' : val }))
            }
          >
            <SelectTrigger className='h-9 w-full sm:w-36 text-xs'>
              <SelectValue placeholder='Semua Arah' />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value='ALL'>Semua Arah</SelectItem>
              <SelectItem value='IN'>Pemasukan (IN)</SelectItem>
              <SelectItem value='OUT'>Pengeluaran (OUT)</SelectItem>
            </SelectContent>
          </Select>
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
                  Memuat data transaksi kas...
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
                  Belum ada mutasi transaksi kas
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
