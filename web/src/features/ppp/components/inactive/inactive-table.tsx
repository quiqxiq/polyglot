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
import type { PPPSecret } from '@/gen/v1/ppp_pb'
import { inactiveColumns } from './inactive-columns'

interface InactiveTableProps {
  data: PPPSecret[]
  isLoading?: boolean
}

export function InactiveTable({ data, isLoading }: InactiveTableProps) {
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])
  const [sorting, setSorting] = useState<SortingState>([])
  const [globalFilter, setGlobalFilter] = useState('')

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
    state: {
      sorting,
      columnVisibility,
      columnFilters,
      globalFilter,
    },
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    onColumnVisibilityChange: setColumnVisibility,
    onGlobalFilterChange: setGlobalFilter,
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
