import { type ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
import { DataTableColumnHeader } from '@/components/data-table'
import type { HotspotIPBinding } from '@/gen/v1/hotspot_pb'
import { BindingsRowActions } from './bindings-row-actions'

export const bindingsColumns: ColumnDef<HotspotIPBinding>[] = [
  {
    id: 'select',
    header: ({ table }) => (
      <Checkbox
        checked={
          table.getIsAllPageRowsSelected() ||
          (table.getIsSomePageRowsSelected() && 'indeterminate')
        }
        onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
        aria-label='Select all'
        className='translate-y-[2px]'
      />
    ),
    cell: ({ row }) => (
      <Checkbox
        checked={row.getIsSelected()}
        onCheckedChange={(value) => row.toggleSelected(!!value)}
        aria-label='Select row'
        className='translate-y-[2px]'
      />
    ),
    enableSorting: false,
    enableHiding: false,
  },
  {
    accessorKey: 'macAddress',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='MAC Address' />
    ),
    cell: ({ row }) => {
      const mac = row.getValue('macAddress') as string
      return <span className='font-mono text-xs font-semibold'>{mac || '-'}</span>
    },
  },
  {
    accessorKey: 'address',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Address' />
    ),
    cell: ({ row }) => {
      const addr = row.getValue('address') as string
      return <span className='font-mono text-xs'>{addr || '-'}</span>
    },
  },
  {
    accessorKey: 'toAddress',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='To Address' />
    ),
    cell: ({ row }) => {
      const toAddr = row.getValue('toAddress') as string
      return <span className='font-mono text-xs text-muted-foreground'>{toAddr || '-'}</span>
    },
  },
  {
    accessorKey: 'server',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Server' />
    ),
    cell: ({ row }) => {
      const server = row.getValue('server') as string
      return (
        <Badge variant='outline' className='text-[10px] font-mono'>
          {server || 'all'}
        </Badge>
      )
    },
  },
  {
    accessorKey: 'type',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Type' />
    ),
    cell: ({ row }) => {
      const type = (row.getValue('type') as string) || 'bypassed'
      let variant: 'default' | 'destructive' | 'secondary' = 'default'
      if (type === 'blocked') variant = 'destructive'
      else if (type === 'regular') variant = 'secondary'

      return (
        <Badge variant={variant} className='uppercase text-[10px] font-semibold tracking-wider'>
          {type}
        </Badge>
      )
    },
    filterFn: (row, id, value: string[]) => {
      return value.includes(row.getValue(id))
    },
  },
  {
    accessorKey: 'comment',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Comment' />
    ),
    cell: ({ row }) => {
      const comment = row.getValue('comment') as string
      if (!comment) return <span className='text-xs text-muted-foreground'>-</span>
      return <span className='text-xs text-muted-foreground truncate max-w-[150px] inline-block'>{comment}</span>
    },
  },
  {
    accessorKey: 'disabled',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Status' />
    ),
    cell: ({ row }) => {
      const disabled = row.getValue('disabled') as boolean
      return (
        <Badge
          variant={disabled ? 'outline' : 'secondary'}
          className={`text-[10px] font-medium ${
            disabled ? 'text-muted-foreground' : 'text-emerald-600 bg-emerald-50 dark:bg-emerald-950/30'
          }`}
        >
          {disabled ? 'Disabled' : 'Enabled'}
        </Badge>
      )
    },
  },
  {
    id: 'actions',
    cell: ({ row }) => <BindingsRowActions row={row} />,
  },
]
