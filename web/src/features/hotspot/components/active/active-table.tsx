import { useState } from 'react'
import {
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
import { Cross2Icon } from '@radix-ui/react-icons'
import { Activity } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { DataTablePagination } from '@/components/data-table'
import type { HotspotActiveSession } from '@/gen/v1/hotspot_pb'
import { ActiveCard } from './active-card'
import { activeColumns as columns } from './active-columns'

interface ActiveTableProps {
  data: HotspotActiveSession[]
  isLoading?: boolean
}

export function ActiveTable({ data, isLoading }: ActiveTableProps) {
  const [sorting, setSorting] = useState<SortingState>([])
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})
  const [globalFilter, setGlobalFilter] = useState('')

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
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
  })

  return (
    <div className='flex flex-1 flex-col gap-4'>
      <div className='flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between'>
        <div className='relative w-full sm:w-72'>
          <Input
            placeholder='Filter active users by user, IP, MAC...'
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

        <div className='flex items-center gap-2'>
          <Badge variant='outline' className='gap-1.5 py-1 px-2.5 text-emerald-600 dark:text-emerald-400 border-emerald-500/30 bg-emerald-500/10'>
            <Activity className='size-3.5 animate-pulse' />
            Live Streaming Native
          </Badge>
        </div>
      </div>

      {/* ===== 1. Mobile Card List View (< md screen) ===== */}
      <div className="space-y-2.5 block md:hidden">
        {isLoading ? (
          <div className="rounded-lg border border-dashed p-8 text-center text-xs text-muted-foreground">
            Connecting real-time active sessions stream...
          </div>
        ) : table.getRowModel().rows.length ? (
          table.getRowModel().rows.map((row) => (
            <ActiveCard key={row.id} session={row.original} />
          ))
        ) : (
          <div className="rounded-lg border border-dashed p-8 text-center text-xs text-muted-foreground">
            No active hotspot sessions online right now.
          </div>
        )}
      </div>

      {/* ===== 2. Desktop Table View (>= md screen) ===== */}
      <div className='overflow-hidden rounded-md border bg-card hidden md:block'>
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
                  Connecting real-time active sessions stream...
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
                  No active hotspot sessions online right now.
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
