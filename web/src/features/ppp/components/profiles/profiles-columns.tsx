import { Badge } from '@/components/ui/badge'
import { DataTableColumnHeader } from '@/components/data-table'
import type { PPPProfile } from '@/gen/v1/ppp_pb'
import type { ColumnDef } from '@tanstack/react-table'
import { Globe, Layers, Users } from 'lucide-react'
import { ProfilesRowActions } from './profiles-row-actions'

export const profilesColumns: ColumnDef<PPPProfile>[] = [
  {
    accessorKey: 'name',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title="Profile Name" />
    ),
    cell: ({ row }) => {
      const name = row.getValue('name') as string
      return (
        <div className="flex items-center gap-2 font-medium">
          <div className="flex h-7 w-7 items-center justify-center rounded bg-primary/10 text-primary">
            <Layers className="h-3.5 w-3.5" />
          </div>
          <span className="font-mono text-sm font-semibold">{name}</span>
        </div>
      )
    },
    enableSorting: true,
  },
  {
    accessorKey: 'rateLimit',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title="Rate Limit (Rx/Tx)" />
    ),
    cell: ({ row }) => {
      const rateLimit = row.getValue('rateLimit') as string
      if (!rateLimit) {
        return <span className="text-xs text-muted-foreground font-mono">Unlimited</span>
      }
      return (
        <Badge
          variant="outline"
          className="border-primary/30 bg-primary/5 text-primary font-mono text-xs font-semibold"
        >
          {rateLimit}
        </Badge>
      )
    },
    enableSorting: true,
  },
  {
    id: 'addresses',
    header: 'Local / Remote Address',
    cell: ({ row }) => {
      const local = row.original.localAddress
      const remote = row.original.remoteAddress
      if (!local && !remote) {
        return <span className="text-xs text-muted-foreground">Default pool</span>
      }
      return (
        <div className="flex flex-col text-xs font-mono">
          {remote && <span>Remote: {remote}</span>}
          {local && <span className="text-muted-foreground">Local: {local}</span>}
        </div>
      )
    },
  },
  {
    accessorKey: 'dnsServer',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title="DNS Servers" />
    ),
    cell: ({ row }) => {
      const dns = row.getValue('dnsServer') as string
      if (!dns) return <span className="text-xs text-muted-foreground">-</span>
      return (
        <div className="flex items-center gap-1.5 text-xs font-mono text-muted-foreground">
          <Globe className="h-3.5 w-3.5" />
          <span>{dns}</span>
        </div>
      )
    },
  },
  {
    accessorKey: 'onlyOne',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title="Only-One" />
    ),
    cell: ({ row }) => {
      const onlyOne = row.getValue('onlyOne') as string
      if (onlyOne === 'yes') {
        return (
          <Badge
            variant="outline"
            className="border-emerald-500/30 bg-emerald-500/10 text-emerald-600 text-[10px]"
          >
            1 Session Only
          </Badge>
        )
      }
      return (
        <span className="text-xs text-muted-foreground font-mono">
          {onlyOne || 'default'}
        </span>
      )
    },
  },
  {
    accessorKey: 'sharedUsers',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title="Shared Users" />
    ),
    cell: ({ row }) => {
      const shared = row.getValue('sharedUsers') as string
      return (
        <div className="flex items-center gap-1 text-xs text-muted-foreground font-mono">
          <Users className="h-3.5 w-3.5" />
          <span>{shared || '1'}</span>
        </div>
      )
    },
  },
  {
    accessorKey: 'comment',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title="Comment" />
    ),
    cell: ({ row }) => {
      const comment = row.getValue('comment') as string
      if (!comment) return <span className="text-xs text-muted-foreground">-</span>
      return (
        <span className="max-w-[160px] truncate text-xs text-muted-foreground">
          {comment}
        </span>
      )
    },
  },
  {
    id: 'actions',
    cell: ({ row }) => <ProfilesRowActions row={row.original} />,
  },
]
