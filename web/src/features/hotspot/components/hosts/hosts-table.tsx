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
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { DataTablePagination } from '@/components/data-table'
import type { HotspotHost } from '@/gen/v1/hotspot_pb'
import { hostsColumns as columns } from './hosts-columns'

interface HostsTableProps {
  data: HotspotHost[]
  isLoading?: boolean
}

type HostFlagFilter = 'all' | 'authorized' | 'bypassed'

export function HostsTable({ data, isLoading }: HostsTableProps) {
  const [sorting, setSorting] = useState<SortingState>([])
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})
  const [globalFilter, setGlobalFilter] = useState('')
  const [flagFilter, setFlagFilter] = useState<HostFlagFilter>('all')

  const filteredData = data.filter((h) => {
    if (flagFilter === 'authorized') return h.authorized
    if (flagFilter === 'bypassed') return h.bypassed
    return true
  })

  const table = useReactTable({
    data: filteredData,
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
        <div className='flex flex-1 flex-wrap items-center gap-2'>
          <div className='relative w-full sm:w-72'>
            <Input
              placeholder='Filter hosts by MAC, IP, comment...'
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

          <div className='flex items-center rounded-md border p-0.5 bg-muted/40'>
            <Button
              variant={flagFilter === 'all' ? 'secondary' : 'ghost'}
              size='sm'
              className='h-7 text-xs px-2.5'
              onClick={() => setFlagFilter('all')}
            >
              All ({data.length})
            </Button>
            <Button
              variant={flagFilter === 'authorized' ? 'secondary' : 'ghost'}
              size='sm'
              className='h-7 text-xs px-2.5 text-emerald-600 dark:text-emerald-400'
              onClick={() => setFlagFilter('authorized')}
            >
              Authorized (A)
            </Button>
            <Button
              variant={flagFilter === 'bypassed' ? 'secondary' : 'ghost'}
              size='sm'
              className='h-7 text-xs px-2.5 text-sky-600 dark:text-sky-400'
              onClick={() => setFlagFilter('bypassed')}
            >
              Bypassed (P)
            </Button>
          </div>
        </div>
      </div>

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
                  Loading hotspot hosts from MikroTik...
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
                  No hotspot host entries found.
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
