import { type ColumnDef } from '@tanstack/react-table'
import { Trash2 } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { DataTableColumnHeader } from '@/components/data-table'
import type { HotspotCookie } from '@/gen/v1/hotspot_pb'
import { useHotspot } from '../../context/hotspot-context'

function CookieActions({ cookie }: { cookie: HotspotCookie }) {
  const { setOpen, setCurrentCookie } = useHotspot()

  return (
    <Button
      variant='ghost'
      size='icon'
      onClick={() => {
        setCurrentCookie(cookie)
        setOpen('cookie-delete')
      }}
      className='size-8 text-destructive hover:text-destructive'
      title='Remove cookie'
    >
      <Trash2 className='size-3.5' />
    </Button>
  )
}

export const cookiesColumns: ColumnDef<HotspotCookie>[] = [
  {
    accessorKey: 'user',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='User / Voucher' />
    ),
    cell: ({ row }) => {
      const user = row.getValue('user') as string
      return <span className='font-semibold text-xs'>{user || '-'}</span>
    },
  },
  {
    accessorKey: 'macAddress',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='MAC Address' />
    ),
    cell: ({ row }) => {
      const mac = row.getValue('macAddress') as string
      return <span className='font-mono text-xs text-muted-foreground'>{mac || '-'}</span>
    },
  },
  {
    accessorKey: 'expiresIn',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Expires In' />
    ),
    cell: ({ row }) => {
      const exp = row.getValue('expiresIn') as string
      return (
        <Badge variant='outline' className='font-mono text-[10px]'>
          {exp || '-'}
        </Badge>
      )
    },
  },
  {
    accessorKey: 'domain',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Domain' />
    ),
    cell: ({ row }) => {
      const domain = row.getValue('domain') as string
      return <span className='text-xs text-muted-foreground'>{domain || '-'}</span>
    },
  },
  {
    id: 'actions',
    cell: ({ row }) => <CookieActions cookie={row.original} />,
  },
]
