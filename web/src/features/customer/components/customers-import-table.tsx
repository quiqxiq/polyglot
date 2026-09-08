import { useState, useMemo, useEffect } from 'react'
import {
  type SortingState,
  type ColumnFiltersState,
  type RowSelectionState,
  type PaginationState,
  type OnChangeFn,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from '@tanstack/react-table'
import type { CustomerSubscriptionImportRow } from '@/gen/v1/ispadmin_pb'
import type { Plan } from '@/gen/v1/plan_pb'
import { Search, AlertTriangle, Eye, EyeOff } from 'lucide-react'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { DataTablePagination } from '@/components/data-table'
import { getCustomersImportColumns } from './customers-import-columns'

interface CustomersImportTableProps {
  rows: CustomerSubscriptionImportRow[]
  plans: Plan[]
  onUpdateRow: (
    index: number,
    updated: Partial<CustomerSubscriptionImportRow>
  ) => void
  onToggleSelect?: (index: number) => void
  onToggleSelectAllFiltered?: (indices: number[], select: boolean) => void
}

export function CustomersImportTable({
  rows,
  plans,
  onUpdateRow,
}: CustomersImportTableProps) {
  const [globalFilter, setGlobalFilter] = useState('')
  const [sorting, setSorting] = useState<SortingState>([])
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])
  const [pagination, setPagination] = useState<PaginationState>({
    pageIndex: 0,
    pageSize: 10,
  })
  const [showPasswords, setShowPasswords] = useState(false)
  const [warningOnly, setWarningOnly] = useState(false)

  const columns = useMemo(
    () => getCustomersImportColumns({ plans, showPasswords, onUpdateRow }),
    [plans, showPasswords, onUpdateRow]
  )

  const rowSelection = useMemo(() => {
    const sel: Record<string, boolean> = {}
    rows.forEach((r, idx) => {
      if (r.selected) sel[idx] = true
    })
    return sel
  }, [rows])

  const handleRowSelectionChange: OnChangeFn<RowSelectionState> = (
    updaterOrValue
  ) => {
    const nextSelection =
      typeof updaterOrValue === 'function'
        ? updaterOrValue(rowSelection)
        : updaterOrValue

    rows.forEach((_, idx) => {
      const isSelected = !!nextSelection[idx]
      if (rows[idx].selected !== isSelected) {
        onUpdateRow(idx, { selected: isSelected })
      }
    })
  }

  const table = useReactTable({
    data: rows,
    columns,
    state: {
      sorting,
      columnFilters,
      globalFilter,
      rowSelection,
      pagination,
    },
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    onGlobalFilterChange: setGlobalFilter,
    onRowSelectionChange: handleRowSelectionChange,
    onPaginationChange: setPagination,
    autoResetPageIndex: false,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
  })

  // Sinkronisasi filter status peringatan
  useEffect(() => {
    table
      .getColumn('status')
      ?.setFilterValue(warningOnly ? 'warning' : undefined)
  }, [warningOnly, table])

  const [typeFilter, setTypeFilter] = useState<
    'ALL' | 'PPPOE' | 'HOTSPOT' | 'IP_BINDING'
  >('ALL')

  const handleTypeFilter = (t: 'ALL' | 'PPPOE' | 'HOTSPOT' | 'IP_BINDING') => {
    setTypeFilter(t)
    setPagination((prev) => ({ ...prev, pageIndex: 0 }))
    if (t === 'ALL') {
      table.getColumn('username')?.setFilterValue(undefined)
    } else if (t === 'IP_BINDING') {
      table.getColumn('username')?.setFilterValue('IP_BINDING')
    } else {
      table.getColumn('username')?.setFilterValue(t)
    }
  }

  return (
    <div className='space-y-4'>
      {/* Top Filter Bar */}
      <div className='flex flex-wrap items-center justify-between gap-3 rounded-lg border bg-card p-3 shadow-xs'>
        <div className='flex flex-wrap items-center gap-2'>
          <div className='relative w-64'>
            <Search className='absolute top-2.5 left-2.5 h-4 w-4 text-muted-foreground' />
            <Input
              placeholder='Cari username, nama, no HP...'
              className='h-9 pl-8 text-xs sm:text-sm'
              value={globalFilter}
              onChange={(e) => {
                setGlobalFilter(e.target.value)
                setPagination((prev) => ({ ...prev, pageIndex: 0 }))
              }}
            />
          </div>

          {/* Type Filter Buttons */}
          <div className='flex items-center gap-1 rounded-md border p-1 text-xs'>
            {[
              { id: 'ALL', label: 'Semua Tipe' },
              { id: 'PPPOE', label: 'PPPoE' },
              { id: 'HOTSPOT', label: 'Hotspot' },
              { id: 'IP_BINDING', label: 'IP Binding' },
            ].map((t) => (
              <button
                key={t.id}
                type='button'
                onClick={() => handleTypeFilter(t.id as typeof typeFilter)}
                className={`rounded px-2.5 py-1 font-medium transition-colors ${
                  typeFilter === t.id
                    ? 'bg-primary text-primary-foreground'
                    : 'text-muted-foreground hover:bg-muted'
                }`}
              >
                {t.label}
              </button>
            ))}
          </div>

          {/* Warning Toggle */}
          <div className='flex items-center gap-1 rounded-md border p-1 text-xs'>
            <button
              type='button'
              onClick={() => {
                setWarningOnly(false)
                setPagination((prev) => ({ ...prev, pageIndex: 0 }))
              }}
              className={`rounded px-2.5 py-1 font-medium transition-colors ${
                !warningOnly
                  ? 'bg-primary text-primary-foreground'
                  : 'text-muted-foreground hover:bg-muted'
              }`}
            >
              Semua Akun
            </button>
            <button
              type='button'
              onClick={() => {
                setWarningOnly(true)
                setPagination((prev) => ({ ...prev, pageIndex: 0 }))
              }}
              className={`flex items-center gap-1 rounded px-2.5 py-1 font-medium transition-colors ${
                warningOnly
                  ? 'bg-amber-500 text-white'
                  : 'text-muted-foreground hover:bg-muted'
              }`}
            >
              <AlertTriangle className='h-3 w-3' />
              Perlu No HP / Alamat
            </button>
          </div>
        </div>

        {/* Toggle Password Visibility */}
        <button
          type='button'
          onClick={() => setShowPasswords(!showPasswords)}
          className='flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground'
        >
          {showPasswords ? (
            <>
              <EyeOff className='h-3.5 w-3.5' /> Sembunyikan Password
            </>
          ) : (
            <>
              <Eye className='h-3.5 w-3.5' /> Lihat Password Router
            </>
          )}
        </button>
      </div>

      {/* TanStack Table View */}
      <div className='overflow-x-auto rounded-md border bg-card'>
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id} className='bg-muted/40'>
                {headerGroup.headers.map((header) => (
                  <TableHead key={header.id}>
                    {header.isPlaceholder
                      ? null
                      : flexRender(
                          header.column.columnDef.header,
                          header.getContext()
                        )}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={columns.length}
                  className='h-24 text-center text-muted-foreground'
                >
                  Tidak ada data pelanggan yang sesuai dengan filter.
                </TableCell>
              </TableRow>
            ) : (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  className={row.getIsSelected() ? 'bg-primary/5' : undefined}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext()
                      )}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {/* Standard DataTable Pagination */}
      <DataTablePagination table={table} />
    </div>
  )
}
