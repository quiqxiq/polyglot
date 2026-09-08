import { useEffect, useState } from 'react'
import {
  type ColumnFiltersState,
  type PaginationState,
  type RowSelectionState,
  type SortingState,
  type VisibilityState,
  getCoreRowModel,
  getFacetedRowModel,
  getFacetedUniqueValues,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from '@tanstack/react-table'
import { DataTablePagination, DataTableToolbar } from '@/components/data-table'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { flexRender } from '@tanstack/react-table'
import type { HotspotCookie } from '@/gen/v1/hotspot_pb'
import { CookiesCard } from './cookies-card'
import { cookiesColumns } from './cookies-columns'

type CookiesTableProps = {
  data: HotspotCookie[]
  isLoading?: boolean
}

export function CookiesTable({ data, isLoading }: CookiesTableProps) {
  const [rowSelection, setRowSelection] = useState<RowSelectionState>({})
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])
  const [sorting, setSorting] = useState<SortingState>([])
  const [pagination, setPagination] = useState<PaginationState>({
    pageIndex: 0,
    pageSize: 10,
  })

  const table = useReactTable({
    data,
    columns: cookiesColumns,
    autoResetPageIndex: false,
    state: {
      sorting,
      columnVisibility,
      rowSelection,
      columnFilters,
      pagination,
    },
    enableRowSelection: true,
    onRowSelectionChange: setRowSelection,
    onSortingChange: setSorting,
    onPaginationChange: setPagination,
    onColumnFiltersChange: (updater) => {
      setColumnFilters(updater)
      setPagination((prev) => ({ ...prev, pageIndex: 0 }))
    },
    onColumnVisibilityChange: setColumnVisibility,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
  })

  // Prevent empty page when deleting last item on current page
  useEffect(() => {
    const pageCount = table.getPageCount()
    if (pageCount > 0 && pagination.pageIndex >= pageCount) {
      setPagination((prev) => ({ ...prev, pageIndex: Math.max(0, pageCount - 1) }))
    }
  }, [data.length, table, pagination.pageIndex])

  return (
    <div className='space-y-4'>
      <DataTableToolbar
        table={table}
        searchPlaceholder='Filter by user or MAC...'
        searchKey='user'
      />

      {/* ===== 1. Mobile Card List View (< md screen) ===== */}
      <div className="space-y-2.5 block md:hidden">
        {isLoading ? (
          <div className="rounded-lg border border-dashed p-8 text-center text-xs text-muted-foreground">
            Loading Cookies...
          </div>
        ) : table.getRowModel().rows?.length ? (
          table.getRowModel().rows.map((row) => (
            <CookiesCard key={row.id} cookie={row.original} />
          ))
        ) : (
          <div className="rounded-lg border border-dashed p-8 text-center text-xs text-muted-foreground">
            No Hotspot Cookies found.
          </div>
        )}
      </div>

      {/* ===== 2. Desktop Table View (>= md screen) ===== */}
      <div className='rounded-md border hidden md:block'>
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead key={header.id} colSpan={header.colSpan}>
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
            {isLoading ? (
              <TableRow>
                <TableCell
                  colSpan={cookiesColumns.length}
                  className='h-24 text-center text-muted-foreground'
                >
                  Loading Cookies...
                </TableCell>
              </TableRow>
            ) : table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && 'selected'}
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
            ) : (
              <TableRow>
                <TableCell
                  colSpan={cookiesColumns.length}
                  className='h-24 text-center text-muted-foreground'
                >
                  No Hotspot Cookies found.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      <DataTablePagination table={table} />
    </div>
  )
}
