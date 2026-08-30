import { useState, useMemo } from 'react'
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { DataTablePagination } from '@/components/data-table'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import type { Registration } from '@/gen/v1/registration_pb'
import { useRegistrationColumns } from './registration-columns'
import { REGISTRATION_STATUS } from '../data/constants'

interface RegistrationTableProps {
  data: Registration[]
  isLoading?: boolean
}

export function RegistrationTable({ data, isLoading }: RegistrationTableProps) {
  const [sorting, setSorting] = useState<SortingState>([])
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})
  const [globalFilter, setGlobalFilter] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('ALL')

  const columns = useRegistrationColumns()

  const filteredData = useMemo(() => {
    if (statusFilter === 'ALL') return data
    if (statusFilter === 'CLOSED') {
      return data.filter(
        (r) =>
          r.status === REGISTRATION_STATUS.REJECTED ||
          r.status === REGISTRATION_STATUS.CANCELLED
      )
    }
    return data.filter((r) => r.status === statusFilter)
  }, [data, statusFilter])

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

  const countByStatus = useMemo(() => {
    const counts: Record<string, number> = {
      ALL: data.length,
      PENDING: 0,
      APPROVED: 0,
      INSTALLED: 0,
      ACTIVE: 0,
      CLOSED: 0,
    }
    for (const item of data) {
      if (item.status === REGISTRATION_STATUS.PENDING) counts.PENDING++
      else if (item.status === REGISTRATION_STATUS.APPROVED) counts.APPROVED++
      else if (item.status === REGISTRATION_STATUS.INSTALLED) counts.INSTALLED++
      else if (item.status === REGISTRATION_STATUS.ACTIVE) counts.ACTIVE++
      else if (
        item.status === REGISTRATION_STATUS.REJECTED ||
        item.status === REGISTRATION_STATUS.CANCELLED
      ) {
        counts.CLOSED++
      }
    }
    return counts
  }, [data])

  return (
    <div className='flex flex-1 flex-col gap-4'>
      <div className='flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between'>
        <Tabs
          value={statusFilter}
          onValueChange={setStatusFilter}
          className='w-full sm:w-auto'
        >
          <TabsList className='grid grid-cols-3 sm:flex sm:h-9'>
            <TabsTrigger value='ALL' className='text-xs'>
              Semua ({countByStatus.ALL})
            </TabsTrigger>
            <TabsTrigger value='PENDING' className='text-xs'>
              Pending ({countByStatus.PENDING})
            </TabsTrigger>
            <TabsTrigger value='APPROVED' className='text-xs'>
              Jadwal Pasang ({countByStatus.APPROVED})
            </TabsTrigger>
            <TabsTrigger value='INSTALLED' className='text-xs'>
              Terpasang ({countByStatus.INSTALLED})
            </TabsTrigger>
            <TabsTrigger value='ACTIVE' className='text-xs'>
              Aktif ({countByStatus.ACTIVE})
            </TabsTrigger>
            <TabsTrigger value='CLOSED' className='text-xs'>
              Ditolak/Batal ({countByStatus.CLOSED})
            </TabsTrigger>
          </TabsList>
        </Tabs>

        <div className='relative w-full sm:w-72'>
          <Input
            placeholder='Cari calon pelanggan...'
            value={globalFilter ?? ''}
            onChange={(e) => setGlobalFilter(e.target.value)}
            className='h-9 pr-8 text-sm'
          />
          {globalFilter && (
            <button
              type='button'
              onClick={() => setGlobalFilter('')}
              className='absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground'
              aria-label='Clear search'
            >
              <Cross2Icon className='h-4 w-4' />
            </button>
          )}
        </div>
      </div>

      <div className='overflow-hidden rounded-md border'>
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead
                    key={header.id}
                    colSpan={header.colSpan}
                    className={(header.column.columnDef.meta as { className?: string })?.className}
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
                  className='h-32 text-center text-muted-foreground'
                >
                  Memuat data pendaftaran...
                </TableCell>
              </TableRow>
            ) : table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && 'selected'}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell
                      key={cell.id}
                      className={(cell.column.columnDef.meta as { className?: string })?.className}
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
                  className='h-32 text-center text-muted-foreground'
                >
                  Tidak ada data pendaftaran calon pelanggan.
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
