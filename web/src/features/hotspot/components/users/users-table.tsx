import { useState, useMemo } from 'react'
import {
  type SortingState,
  type VisibilityState,
  type ColumnFiltersState,
  type PaginationState,
  flexRender,
  getCoreRowModel,
  getFacetedRowModel,
  getFacetedUniqueValues,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from '@tanstack/react-table'
import { Printer } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { DataTablePagination, DataTableToolbar } from '@/components/data-table'
import type { HotspotUser, HotspotProfile } from '@/gen/v1/hotspot_pb'
import { UsersCard } from './users-card'
import { usersColumns as columns } from './users-columns'
import { UsersBulkActions } from './users-bulk-actions'
import { useHotspot } from '../../context/hotspot-context'
import { cn } from '@/lib/utils'

interface UsersTableProps {
  data: HotspotUser[]
  profiles: HotspotProfile[]
  isLoading?: boolean
}

export function UsersTable({
  data,
  profiles,
  isLoading,
}: UsersTableProps) {
  const { setOpen, setPrintBatchComment, setPrintSingleUserId } = useHotspot()
  const [rowSelection, setRowSelection] = useState({})
  const [sorting, setSorting] = useState<SortingState>([])
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])
  const [globalFilter, setGlobalFilter] = useState('')
  const [pagination, setPagination] = useState<PaginationState>({
    pageIndex: 0,
    pageSize: 10,
  })

  // Derive profile options for faceted filter
  const profileOptions = useMemo(() => {
    const profileNames = new Set<string>()
    profiles.forEach((p) => p.name && profileNames.add(p.name))
    data.forEach((u) => u.profile && profileNames.add(u.profile))
    return Array.from(profileNames).map((name) => ({
      label: name,
      value: name,
    }))
  }, [profiles, data])

  // Derive batch/comment options for faceted filter
  const commentOptions = useMemo(() => {
    const comments = Array.from(
      new Set(data.map((u) => u.comment?.trim()).filter(Boolean))
    ) as string[]
    return comments.map((c) => ({
      label: c,
      value: c,
    }))
  }, [data])

  const table = useReactTable({
    data,
    columns,
    state: {
      sorting,
      columnVisibility,
      rowSelection,
      columnFilters,
      globalFilter,
      pagination,
    },
    enableRowSelection: true,
    onRowSelectionChange: setRowSelection,
    onSortingChange: setSorting,
    onColumnVisibilityChange: setColumnVisibility,
    onColumnFiltersChange: setColumnFilters,
    onGlobalFilterChange: setGlobalFilter,
    onPaginationChange: setPagination,
    globalFilterFn: (row, _columnId, filterValue) => {
      const name = String(row.getValue('name') || '').toLowerCase()
      const comment = String(row.getValue('comment') || '').toLowerCase()
      const profile = String(row.getValue('profile') || '').toLowerCase()
      const server = String(row.getValue('server') || '').toLowerCase()
      const searchValue = String(filterValue).toLowerCase()

      return (
        name.includes(searchValue) ||
        comment.includes(searchValue) ||
        profile.includes(searchValue) ||
        server.includes(searchValue)
      )
    },
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
  })

  // Detect if a single batch/comment is filtered to offer a quick print button
  const commentFilterValues =
    (table.getColumn('comment')?.getFilterValue() as string[]) || []
  const activeBatch =
    commentFilterValues.length === 1 ? commentFilterValues[0] : null

  const handlePrintBatch = (batchTag: string) => {
    setPrintBatchComment(batchTag)
    setPrintSingleUserId('')
    setOpen('voucher-print')
  }

  return (
    <div
      className={cn(
        'max-sm:has-[div[role="toolbar"]]:mb-16',
        'flex flex-1 flex-col gap-4'
      )}
    >
      {/* ===== Toolbar with Search & Faceted Filters ===== */}
      <div className='flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between'>
        <div className='flex-1'>
          <DataTableToolbar
            table={table}
            searchPlaceholder='Filter by username, profile, comment...'
            filters={[
              {
                columnId: 'profile',
                title: 'Profile',
                options: profileOptions,
              },
              {
                columnId: 'comment',
                title: 'Batch / Comment',
                options: commentOptions,
              },
            ]}
          />
        </div>

        {activeBatch && (
          <Button
            size='sm'
            variant='outline'
            onClick={() => handlePrintBatch(activeBatch)}
            className='h-8 gap-1.5 text-primary border-primary hover:bg-primary/10 shrink-0 self-start sm:self-auto'
          >
            <Printer className='size-4' />
            Print Batch ({activeBatch})
          </Button>
        )}
      </div>

      {/* ===== 1. Mobile Card List View (< md screen) ===== */}
      <div className="space-y-2.5 block md:hidden">
        {isLoading ? (
          <div className="rounded-lg border border-dashed p-8 text-center text-xs text-muted-foreground">
            Loading hotspot users from MikroTik...
          </div>
        ) : table.getRowModel().rows?.length ? (
          table.getRowModel().rows.map((row) => (
            <UsersCard key={row.id} user={row.original} />
          ))
        ) : (
          <div className="rounded-lg border border-dashed p-8 text-center text-xs text-muted-foreground">
            No hotspot users found.
          </div>
        )}
      </div>

      {/* ===== 2. Desktop Table View (>= md screen) ===== */}
      <div className='overflow-hidden rounded-md border bg-card hidden md:block'>
        <Table className='min-w-xl'>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead
                    key={header.id}
                    colSpan={header.colSpan}
                    className={cn(
                      header.column.columnDef.meta?.className,
                      header.column.columnDef.meta?.thClassName
                    )}
                  >
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
                  colSpan={columns.length}
                  className='h-24 text-center text-muted-foreground'
                >
                  Loading hotspot users from MikroTik...
                </TableCell>
              </TableRow>
            ) : table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && 'selected'}
                  className={row.original.disabled ? 'opacity-60 bg-muted/30 text-muted-foreground' : ''}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell
                      key={cell.id}
                      className={cn(
                        cell.column.columnDef.meta?.className,
                        cell.column.columnDef.meta?.tdClassName
                      )}
                    >
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
                  colSpan={columns.length}
                  className='h-24 text-center text-muted-foreground'
                >
                  No hotspot users found.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      <DataTablePagination table={table} className='mt-auto' />
      <UsersBulkActions table={table} />
    </div>
  )
}
