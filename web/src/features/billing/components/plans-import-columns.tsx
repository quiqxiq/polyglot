import type { ColumnDef } from '@tanstack/react-table'
import type { PlanImportRow } from '@/gen/v1/ispadmin_pb'
import { Wand2 } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { DataTableColumnHeader } from '@/components/data-table'
import { formatAutoName } from '@/lib/format-name'

interface GetPlansImportColumnsProps {
  onUpdateRow: (index: number, updated: Partial<PlanImportRow>) => void
}

export function getPlansImportColumns({
  onUpdateRow,
}: GetPlansImportColumnsProps): ColumnDef<PlanImportRow>[] {
  return [
    {
      id: 'select',
      header: ({ table }) => (
        <div className='flex justify-center'>
          <Checkbox
            checked={
              table.getIsAllPageRowsSelected() ||
              (table.getIsSomePageRowsSelected() && 'indeterminate')
            }
            onCheckedChange={(value) =>
              table.toggleAllPageRowsSelected(!!value)
            }
            aria-label='Pilih semua di halaman ini'
          />
        </div>
      ),
      cell: ({ row }) => (
        <div className='flex justify-center'>
          <Checkbox
            checked={row.getIsSelected()}
            onCheckedChange={(value) => row.toggleSelected(!!value)}
            aria-label={`Pilih paket ${row.original.name}`}
          />
        </div>
      ),
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: 'isNew',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Status' />
      ),
      cell: ({ row }) =>
        row.original.isNew ? (
          <Badge
            variant='outline'
            className='border-emerald-500/20 bg-emerald-500/10 text-xs font-semibold text-emerald-600'
          >
            BARU
          </Badge>
        ) : (
          <Badge
            variant='outline'
            className='border-blue-500/20 bg-blue-500/10 text-xs text-blue-600'
          >
            UPDATE
          </Badge>
        ),
    },
    {
      accessorKey: 'name',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Nama Paket (Komersial)' />
      ),
      cell: ({ row }) => {
        const canFormat =
          row.original.routerProfile &&
          (row.original.routerProfile.includes('_') ||
            row.original.routerProfile.includes('-'))
        return (
          <div className='flex min-w-[180px] flex-col gap-1'>
            <div className='flex items-center gap-1'>
              <Input
                className='h-8 flex-1 text-xs font-medium'
                value={row.original.name}
                onChange={(e) => onUpdateRow(row.index, { name: e.target.value })}
              />
              {canFormat && (
                <Button
                  variant='ghost'
                  size='icon'
                  className='h-7 w-7 shrink-0 text-muted-foreground hover:text-sky-500'
                  title={`Format nama: ${formatAutoName(row.original.routerProfile)}`}
                  onClick={() =>
                    onUpdateRow(row.index, {
                      name: formatAutoName(row.original.routerProfile),
                    })
                  }
                >
                  <Wand2 className='h-3.5 w-3.5 text-sky-500' />
                </Button>
              )}
            </div>
            {row.original.routerProfile && (
              <span className='font-mono text-[10px] text-muted-foreground'>
                Profil Router:{' '}
                <span className='font-semibold text-foreground/80'>
                  {row.original.routerProfile}
                </span>
              </span>
            )}
          </div>
        )
      },
    },
    {
      accessorKey: 'serviceType',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Tipe' />
      ),
      cell: ({ row }) => (
        <Badge
          variant='secondary'
          className={`text-xs ${
            row.original.serviceType === 'PPPOE'
              ? 'bg-sky-500/15 text-sky-700 dark:text-sky-400'
              : 'bg-amber-500/15 text-amber-700 dark:text-amber-400'
          }`}
        >
          {row.original.serviceType}
        </Badge>
      ),
    },
    {
      accessorKey: 'rateLimit',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Rate Limit' />
      ),
      cell: ({ row }) => (
        <span className='font-mono text-xs text-muted-foreground'>
          {row.original.rateLimit || '-'}
        </span>
      ),
    },
    {
      accessorKey: 'bandwidthDownloadKbps',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='DL (Kbps)' />
      ),
      cell: ({ row }) => (
        <Input
          type='number'
          className='h-8 w-24 font-mono text-xs'
          value={row.original.bandwidthDownloadKbps || ''}
          onChange={(e) =>
            onUpdateRow(row.index, {
              bandwidthDownloadKbps: parseInt(e.target.value, 10) || 0,
            })
          }
        />
      ),
    },
    {
      accessorKey: 'bandwidthUploadKbps',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='UL (Kbps)' />
      ),
      cell: ({ row }) => (
        <Input
          type='number'
          className='h-8 w-24 font-mono text-xs'
          value={row.original.bandwidthUploadKbps || ''}
          onChange={(e) =>
            onUpdateRow(row.index, {
              bandwidthUploadKbps: parseInt(e.target.value, 10) || 0,
            })
          }
        />
      ),
    },
    {
      accessorKey: 'price',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Harga Jual (Rp)' />
      ),
      cell: ({ row }) => (
        <div className='relative min-w-[130px]'>
          <span className='absolute top-2 left-2.5 text-xs text-muted-foreground'>
            Rp
          </span>
          <Input
            type='number'
            className={`h-8 pl-8 text-xs font-semibold ${
              row.original.price === 0
                ? 'border-amber-400 bg-amber-500/5'
                : 'border-input'
            }`}
            placeholder='0'
            value={row.original.price || ''}
            onChange={(e) =>
              onUpdateRow(row.index, {
                price: parseFloat(e.target.value) || 0,
              })
            }
          />
        </div>
      ),
    },
    {
      accessorKey: 'ipPoolName',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='IP Pool' />
      ),
      cell: ({ row }) => (
        <Input
          className='h-8 w-28 text-xs'
          value={row.original.ipPoolName}
          placeholder='(none)'
          onChange={(e) =>
            onUpdateRow(row.index, {
              ipPoolName: e.target.value,
            })
          }
        />
      ),
    },
    {
      accessorKey: 'sharedUsers',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Shared' />
      ),
      cell: ({ row }) => (
        <Input
          type='number'
          className='h-8 w-16 text-center text-xs'
          value={row.original.sharedUsers || 1}
          onChange={(e) =>
            onUpdateRow(row.index, {
              sharedUsers: parseInt(e.target.value, 10) || 1,
            })
          }
        />
      ),
    },
  ]
}
