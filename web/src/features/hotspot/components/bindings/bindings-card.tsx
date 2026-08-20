import { Badge } from '@/components/ui/badge'
import type { HotspotIPBinding } from '@/gen/v1/hotspot_pb'
import type { Row } from '@tanstack/react-table'
import { Globe, Network, Server, Shield } from 'lucide-react'
import { BindingsRowActions } from './bindings-row-actions'

interface BindingsCardProps {
  row: Row<HotspotIPBinding>
}

export function BindingsCard({ row }: BindingsCardProps) {
  const binding = row.original
  const mac = binding.macAddress || '-'
  const address = binding.address || '-'
  const toAddress = binding.toAddress || '-'
  const server = binding.server || 'all'
  const type = (binding.type || 'bypassed').toLowerCase()

  let typeBadgeVariant: 'default' | 'destructive' | 'secondary' = 'default'
  if (type === 'blocked') typeBadgeVariant = 'destructive'
  else if (type === 'regular') typeBadgeVariant = 'secondary'

  return (
    <div
      className={`rounded-lg border bg-card p-3.5 shadow-sm space-y-2.5 transition-colors ${
        binding.disabled
          ? 'opacity-60 bg-muted/40 border-muted text-muted-foreground'
          : 'hover:border-primary/40'
      }`}
    >
      {/* Header: Title & Type badge below title, Actions on right */}
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-2.5 min-w-0">
          <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-primary/10 text-primary">
            <Shield className="h-4 w-4" />
          </div>
          <div className="min-w-0 space-y-0.5">
            <div className="font-mono text-sm font-semibold truncate">
              {binding.macAddress || binding.address || 'IP Binding'}
            </div>
            <div>
              <Badge variant={typeBadgeVariant} className="uppercase text-[9px] font-semibold tracking-wider px-1.5 py-0 h-4">
                {type}
              </Badge>
            </div>
          </div>
        </div>

        <div className="shrink-0">
          <BindingsRowActions row={row} />
        </div>
      </div>

      {/* Grid: MAC, Address / To Address, Server */}
      <div className="grid grid-cols-2 gap-2 pt-1 border-t border-border/50 text-xs">
        <div className="space-y-0.5">
          <span className="text-[10px] uppercase tracking-wider text-muted-foreground font-medium">
            MAC Address
          </span>
          <div className="flex items-center gap-1.5 font-mono text-muted-foreground">
            <Network className="h-3.5 w-3.5 shrink-0" />
            <span className="truncate">{mac}</span>
          </div>
        </div>

        <div className="space-y-0.5">
          <span className="text-[10px] uppercase tracking-wider text-muted-foreground font-medium">
            Address / To Address
          </span>
          <div className="flex items-center gap-1.5 font-mono text-muted-foreground">
            <Globe className="h-3.5 w-3.5 text-primary/70 shrink-0" />
            <span className="truncate">{address} {toAddress !== '-' ? `→ ${toAddress}` : ''}</span>
          </div>
        </div>

        <div className="col-span-2 space-y-0.5">
          <span className="text-[10px] uppercase tracking-wider text-muted-foreground font-medium">
            Server
          </span>
          <div className="flex items-center gap-1.5 font-mono text-muted-foreground">
            <Server className="h-3.5 w-3.5 shrink-0" />
            <span className="truncate">{server}</span>
          </div>
        </div>
      </div>
    </div>
  )
}
