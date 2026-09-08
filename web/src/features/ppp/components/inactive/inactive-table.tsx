import { useEffect, useMemo, useState } from 'react'
import {
  type ColumnFiltersState,
  type PaginationState,
  type RowSelectionState,
  type SortingState,
  type VisibilityState,
  flexRender,
  getCoreRowModel,
  getFacetedRowModel,
  getFacetedUniqueValues,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from '@tanstack/react-table'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  DataTablePagination,
  DataTableToolbar,
} from '@/components/data-table'
import type { PPPSecret } from '@/gen/v1/ppp_pb'
import { InactiveCard } from './inactive-card'
import { inactiveColumns } from './inactive-columns'

interface InactiveTableProps {
  data: PPPSecret[]
  isLoading?: boolean
  defaultGlobalFilter?: string
}

export function InactiveTable({ data, isLoading, defaultGlobalFilter }: InactiveTableProps) {
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])
  const [sorting, setSorting] = useState<SortingState>([])
  const [rowSelection, setRowSelection] = useState<RowSelectionState>({})
  const [pagination, setPagination] = useState<PaginationState>({
    pageIndex: 0,
    pageSize: 10,
  })
  const [globalFilter, setGlobalFilter] = useState(defaultGlobalFilter ?? '')

  useEffect(() => {
    if (defaultGlobalFilter !== undefined) {
      setGlobalFilter(defaultGlobalFilter)
    }
  }, [defaultGlobalFilter])


  // Derive unique profiles from inactive data
  const profileOptions = useMemo(() => {
    const map = new Map<string, number>()
    data.forEach((s) => {
      const p = s.profile || 'default'
      map.set(p, (map.get(p) || 0) + 1)
    })
    return Array.from(map.entries()).map(([value, count]) => ({
      label: `${value} (${count})`,
      value,
    }))
  }, [data])

  const table = useReactTable({
    data,
    columns: inactiveColumns,
    autoResetPageIndex: false,
    state: {
      sorting,
      columnVisibility,
      rowSelection,
      columnFilters,
      globalFilter,
      pagination,
    },
    onSortingChange: setSorting,
    onRowSelectionChange: setRowSelection,
    onPaginationChange: setPagination,
    onColumnFiltersChange: setColumnFilters,
    onColumnVisibilityChange: setColumnVisibility,
    onGlobalFilterChange: (updater) => {
      setGlobalFilter(updater)
      setPagination((prev) => ({ ...prev, pageIndex: 0 }))
    },
    globalFilterFn: (row, _, filterValue: string) => {
      const search = filterValue.toLowerCase()
      const name = (row.original.name || '').toLowerCase()
      const callerId = (row.original.callerId || '').toLowerCase()
      const profile = (row.original.profile || '').toLowerCase()
      return (
        name.includes(search) ||
        callerId.includes(search) ||
        profile.includes(search)
      )
    },
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
    <div className="space-y-4">
      <DataTableToolbar
        table={table}
        searchPlaceholder="Search offline subscriber username, MAC, or profile..."
        filters={[
          {
            columnId: 'profile',
            title: 'Profile',
            options: profileOptions,
          },
        ]}
      />

      {/* 1. Mobile Card List View (< md screen) */}
      <div className="space-y-2.5 block md:hidden">
        {table.getRowModel().rows?.length ? (
          table.getRowModel().rows.map((row) => (
            <InactiveCard key={row.id} secret={row.original} />
          ))
        ) : (
          <div className="rounded-lg border border-dashed p-8 text-center text-xs text-muted-foreground">
            {isLoading
              ? 'Loading offline subscribers...'
              : 'All subscribers are currently connected and active!'}
          </div>
        )}
      </div>

      {/* 2. Desktop Table View (>= md screen) */}
      <div className="rounded-md border hidden md:block">
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
            {table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && 'selected'}
                  className={row.original.disabled ? 'opacity-60 bg-muted/30 text-muted-foreground' : ''}
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
                  colSpan={inactiveColumns.length}
                  className="h-24 text-center text-muted-foreground"
                >
                  {isLoading
                    ? 'Loading offline subscribers...'
                    : 'All subscribers are currently connected and active!'}
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
