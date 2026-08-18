import { useMemo, useState } from 'react'
import {
  type ColumnFiltersState,
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
import type { PPPActiveSession } from '@/gen/v1/ppp_pb'
import { activeColumns } from './active-columns'
import { ActiveBulkActions } from './active-bulk-actions'

interface ActiveTableProps {
  data: PPPActiveSession[]
  isLoading?: boolean
}

export function ActiveTable({ data, isLoading }: ActiveTableProps) {
  const [rowSelection, setRowSelection] = useState({})
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])
  const [sorting, setSorting] = useState<SortingState>([])
  const [globalFilter, setGlobalFilter] = useState('')

  // Derive unique services from active data
  const serviceOptions = useMemo(() => {
    const map = new Map<string, number>()
    data.forEach((s) => {
      const srv = (s.service || 'pppoe').toLowerCase()
      map.set(srv, (map.get(srv) || 0) + 1)
    })
    return Array.from(map.entries()).map(([value, count]) => ({
      label: `${value.toUpperCase()} (${count})`,
      value,
    }))
  }, [data])

  // Derive auth types
  const authOptions = useMemo(() => {
    let radius = 0
    let local = 0
    data.forEach((s) => {
      if (s.radius) radius++
      else local++
    })
    return [
      { label: `Local (${local})`, value: 'local' },
      { label: `RADIUS (${radius})`, value: 'radius' },
    ]
  }, [data])

  const table = useReactTable({
    data,
    columns: activeColumns,
    state: {
      sorting,
      columnVisibility,
      rowSelection,
      columnFilters,
      globalFilter,
    },
    enableRowSelection: true,
    onRowSelectionChange: setRowSelection,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    onColumnVisibilityChange: setColumnVisibility,
    onGlobalFilterChange: setGlobalFilter,
    globalFilterFn: (row, _, filterValue: string) => {
      const search = filterValue.toLowerCase()
      const name = (row.original.name || '').toLowerCase()
      const address = (row.original.address || '').toLowerCase()
      const callerId = (row.original.callerId || '').toLowerCase()
      const service = (row.original.service || '').toLowerCase()
      return (
        name.includes(search) ||
        address.includes(search) ||
        callerId.includes(search) ||
        service.includes(search)
      )
    },
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
  })

  return (
    <div className="space-y-4">
      <DataTableToolbar
        table={table}
        searchPlaceholder="Search active session username, IP, or MAC..."
        filters={[
          {
            columnId: 'service',
            title: 'Service',
            options: serviceOptions,
          },
          {
            columnId: 'radius',
            title: 'Auth',
            options: authOptions,
          },
        ]}
      />

      <div className="rounded-md border">
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
                  colSpan={activeColumns.length}
                  className="h-24 text-center text-muted-foreground"
                >
                  {isLoading
                    ? 'Loading active PPPoE sessions...'
                    : 'No active sessions currently connected.'}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      <DataTablePagination table={table} />
      <ActiveBulkActions table={table} />
    </div>
  )
}
