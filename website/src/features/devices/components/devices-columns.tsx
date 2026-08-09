import { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
import { DataTableColumnHeader } from '@/components/data-table'
import { Device } from '@/gen/v1/device_pb'
import { DataTableRowActions } from './data-table-row-actions'

export const devicesColumns: ColumnDef<Device>[] = [
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
        className='translate-y-0.5'
      />
    ),
    cell: ({ row }) => (
      <Checkbox
        checked={row.getIsSelected()}
        onCheckedChange={(value) => row.toggleSelected(!!value)}
        aria-label='Select row'
        className='translate-y-0.5'
      />
    ),
    enableSorting: false,
    enableHiding: false,
  },
  {
    accessorKey: 'name',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Name' />
    ),
    cell: ({ row }) => (
      <div className='font-medium'>{row.getValue('name')}</div>
    ),
  },
  {
    accessorKey: 'host',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Host / Address' />
    ),
    cell: ({ row }) => {
      const device = row.original
      return (
        <div className='font-mono text-xs'>
          {device.host}:{device.port}
        </div>
      )
    },
  },
  {
    accessorKey: 'vendor',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Vendor' />
    ),
    cell: ({ row }) => (
      <Badge variant='outline' className='capitalize'>
        {row.getValue('vendor')}
      </Badge>
    ),
  },
  {
    accessorKey: 'driverType',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Driver Type' />
    ),
    cell: ({ row }) => (
      <div className='text-xs text-muted-foreground uppercase font-mono'>
        {row.getValue('driverType')}
      </div>
    ),
  },
  {
    accessorKey: 'enabled',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Status' />
    ),
    cell: ({ row }) => {
      const enabled = row.getValue('enabled') as boolean
      return (
        <Badge variant={enabled ? 'default' : 'secondary'}>
          {enabled ? 'Active' : 'Disabled'}
        </Badge>
      )
    },
  },
  {
    accessorKey: 'tags',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Tags' />
    ),
    cell: ({ row }) => {
      const tags = (row.getValue('tags') as string[]) || []
      if (tags.length === 0) return <span className='text-xs text-muted-foreground'>—</span>
      return (
        <div className='flex flex-wrap gap-1'>
          {tags.map((tag, i) => (
            <Badge key={i} variant='outline' className='text-[10px] px-1.5 py-0'>
              {tag}
            </Badge>
          ))}
        </div>
      )
    },
  },
  {
    id: 'actions',
    cell: ({ row }) => <DataTableRowActions row={row} />,
  },
]
