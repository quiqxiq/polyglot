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
import { secretsColumns } from './secrets-columns'
import { SecretsBulkActions } from './secrets-bulk-actions'
import { SecretsCard } from './secrets-card'

interface SecretsTableProps {
  data: PPPSecret[]
  isLoading?: boolean
}

export function SecretsTable({ data, isLoading }: SecretsTableProps) {
  const [rowSelection, setRowSelection] = useState({})
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])
  const [sorting, setSorting] = useState<SortingState>([])
  const [globalFilter, setGlobalFilter] = useState('')

  // Derive unique profiles from data for faceted filtering
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

  // Derive unique services from data
  const serviceOptions = useMemo(() => {
    const map = new Map<string, number>()
    data.forEach((s) => {
      const srv = (s.service || 'any').toLowerCase()
      map.set(srv, (map.get(srv) || 0) + 1)
    })
    return Array.from(map.entries()).map(([value, count]) => ({
      label: `${value.toUpperCase()} (${count})`,
      value,
    }))
  }, [data])


  // Derive unique comments / batch tags for faceted filtering
  const commentOptions = useMemo(() => {
    const map = new Map<string, number>()
    data.forEach((s) => {
      if (s.comment) {
        map.set(s.comment, (map.get(s.comment) || 0) + 1)
      }
    })
    return Array.from(map.entries()).map(([value, count]) => ({
      label: `${value} (${count})`,
      value,
    }))
  }, [data])

  const table = useReactTable({
    data,
    columns: secretsColumns,
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
      const comment = (row.original.comment || '').toLowerCase()
      const profile = (row.original.profile || '').toLowerCase()
      const remote = (row.original.remoteAddress || '').toLowerCase()
      const local = (row.original.localAddress || '').toLowerCase()
      return (
        name.includes(search) ||
        comment.includes(search) ||
        profile.includes(search) ||
        remote.includes(search) ||
        local.includes(search)
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
        searchPlaceholder="Search secret username, IP, profile, or comment..."
        filters={[
          {
            columnId: 'profile',
            title: 'Profile',
            options: profileOptions,
          },
          {
            columnId: 'service',
            title: 'Service',
            options: serviceOptions,
          },
          {
            columnId: 'comment',
            title: 'Comment / Batch',
            options: commentOptions,
          },
        ]}
      />

      {/* 1. Mobile Card List View (< md screen) */}
      <div className="space-y-2.5 block md:hidden">
        {table.getRowModel().rows?.length ? (
          table.getRowModel().rows.map((row) => (
            <SecretsCard key={row.id} secret={row.original} />
          ))
        ) : (
          <div className="rounded-lg border border-dashed p-8 text-center text-xs text-muted-foreground">
            {isLoading ? 'Loading PPPoE secrets...' : 'No secrets found.'}
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
                  colSpan={secretsColumns.length}
                  className="h-24 text-center text-muted-foreground"
                >
                  {isLoading ? 'Loading PPPoE secrets...' : 'No secrets found.'}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      <DataTablePagination table={table} />
      <SecretsBulkActions table={table} />
    </div>
  )
}
