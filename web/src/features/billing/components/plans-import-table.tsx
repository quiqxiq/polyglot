import { useState, useMemo } from 'react'
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
import type { PlanImportRow } from '@/gen/v1/ispadmin_pb'
import { Search, Sparkles, Tag, Wand2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
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
import { getPlansImportColumns } from './plans-import-columns'

interface PlansImportTableProps {
  rows: PlanImportRow[]
  onUpdateRow: (index: number, updated: Partial<PlanImportRow>) => void
  onSelectOnlyNew: () => void
  onBatchSetPrice: (price: number) => void
  onAutoFormatNames?: () => void
  onToggleSelect?: (index: number) => void
  onToggleSelectAll?: (selected: boolean) => void
}

export function PlansImportTable({
  rows,
  onUpdateRow,
  onSelectOnlyNew,
  onBatchSetPrice,
  onAutoFormatNames,
}: PlansImportTableProps) {
  const [globalFilter, setGlobalFilter] = useState('')
  const [sorting, setSorting] = useState<SortingState>([])
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])
  const [pagination, setPagination] = useState<PaginationState>({
    pageIndex: 0,
    pageSize: 10,
  })
  const [batchPriceInput, setBatchPriceInput] = useState('')

  const columns = useMemo(
    () => getPlansImportColumns({ onUpdateRow }),
    [onUpdateRow]
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

  const currentTypeFilter =
    (table.getColumn('serviceType')?.getFilterValue() as string) ?? 'ALL'
  const currentStatusFilter = table.getColumn('isNew')?.getFilterValue() as
    | boolean
    | undefined

  const handleTypeFilter = (type: 'ALL' | 'PPPOE' | 'HOTSPOT') => {
    setPagination((prev) => ({ ...prev, pageIndex: 0 }))
    table
      .getColumn('serviceType')
      ?.setFilterValue(type === 'ALL' ? undefined : type)
  }

  const handleStatusFilter = (status: 'ALL' | 'NEW' | 'EXISTING') => {
    setPagination((prev) => ({ ...prev, pageIndex: 0 }))
    if (status === 'ALL') {
      table.getColumn('isNew')?.setFilterValue(undefined)
    } else {
      table.getColumn('isNew')?.setFilterValue(status === 'NEW')
    }
  }

  const handleApplyBatchPrice = () => {
    const num = parseFloat(batchPriceInput.replace(/[^0-9]/g, ''))
    if (!isNaN(num) && num >= 0) {
      onBatchSetPrice(num)
      setBatchPriceInput('')
    }
  }

  return (
    <div className='space-y-4'>
      {/* Top Filter & Batch Toolbar */}
      <div className='flex flex-wrap items-center justify-between gap-3 rounded-lg border bg-card p-3 shadow-xs'>
        <div className='flex flex-wrap items-center gap-2'>
          <div className='relative w-64'>
            <Search className='absolute top-2.5 left-2.5 h-4 w-4 text-muted-foreground' />
            <Input
              placeholder='Cari nama, pool, rate...'
              className='h-9 pl-8 text-xs sm:text-sm'
              value={globalFilter}
              onChange={(e) => {
                setGlobalFilter(e.target.value)
                setPagination((prev) => ({ ...prev, pageIndex: 0 }))
              }}
            />
          </div>

          {/* Filter Tipe Layanan */}
          <div className='flex items-center gap-1 rounded-md border p-1 text-xs'>
            {(['ALL', 'PPPOE', 'HOTSPOT'] as const).map((t) => (
              <button
                key={t}
                type='button'
                onClick={() => handleTypeFilter(t)}
                className={`rounded px-2.5 py-1 font-medium transition-colors ${
                  currentTypeFilter === (t === 'ALL' ? 'ALL' : t)
                    ? 'bg-primary text-primary-foreground'
                    : 'text-muted-foreground hover:bg-muted'
                }`}
              >
                {t === 'ALL' ? 'Semua Tipe' : t}
              </button>
            ))}
          </div>

          {/* Filter Status */}
          <div className='flex items-center gap-1 rounded-md border p-1 text-xs'>
            {[
              { id: 'ALL', label: 'Semua Status' },
              { id: 'NEW', label: 'Paket Baru' },
              { id: 'EXISTING', label: 'Sudah Ada' },
            ].map((s) => (
              <button
                key={s.id}
                type='button'
                onClick={() =>
                  handleStatusFilter(s.id as 'ALL' | 'NEW' | 'EXISTING')
                }
                className={`rounded px-2.5 py-1 font-medium transition-colors ${
                  (s.id === 'ALL' && currentStatusFilter === undefined) ||
                  (s.id === 'NEW' && currentStatusFilter === true) ||
                  (s.id === 'EXISTING' && currentStatusFilter === false)
                    ? 'bg-primary text-primary-foreground'
                    : 'text-muted-foreground hover:bg-muted'
                }`}
              >
                {s.label}
              </button>
            ))}
          </div>
        </div>

        {/* Quick Batch Tools */}
        <div className='flex flex-wrap items-center gap-2'>
          <Button
            variant='outline'
            size='sm'
            onClick={onSelectOnlyNew}
            className='h-8 text-xs'
          >
            <Sparkles className='mr-1 h-3.5 w-3.5 text-amber-500' />
            Pilih Hanya Paket Baru
          </Button>

          {onAutoFormatNames && (
            <Button
              variant='outline'
              size='sm'
              onClick={onAutoFormatNames}
              className='h-8 text-xs'
              title='Format nama profil seperti "10m_unlimited" menjadi "10m Unlimited"'
            >
              <Wand2 className='mr-1 h-3.5 w-3.5 text-sky-500' />
              Auto Format Nama
            </Button>
          )}

          <div className='flex items-center gap-1'>
            <Input
              placeholder='Set harga terpilih...'
              className='h-8 w-36 text-xs'
              value={batchPriceInput}
              onChange={(e) => setBatchPriceInput(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleApplyBatchPrice()}
            />
            <Button
              variant='secondary'
              size='sm'
              onClick={handleApplyBatchPrice}
              className='h-8 text-xs'
            >
              <Tag className='mr-1 h-3 w-3' />
              Terapkan
            </Button>
          </div>
        </div>
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
                  Tidak ada profil paket yang sesuai dengan filter.
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
