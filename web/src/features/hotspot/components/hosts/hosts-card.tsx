import { Badge } from '@/components/ui/badge'
import type { HotspotHost } from '@/gen/v1/hotspot_pb'
import { Globe, Laptop, Server } from 'lucide-react'
import { HostsRowActions } from './hosts-row-actions'

interface HostsCardProps {
  host: HotspotHost
}

export function HostsCard({ host }: HostsCardProps) {
  const ip = host.address || '-'
  const toIp = host.toAddress || '-'
  const server = host.server || 'all'
  const authorized = host.authorized
  const bypassed = host.bypassed

  return (
    <div className="rounded-lg border bg-card p-3.5 shadow-sm space-y-2.5 transition-colors hover:border-primary/40">
      {/* Header: MAC Address & Status badge below MAC, Actions on right */}
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-2.5 min-w-0">
          <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-primary/10 text-primary">
            <Laptop className="h-4 w-4" />
          </div>
          <div className="min-w-0 space-y-0.5">
            <div className="font-mono text-sm font-semibold truncate">
              {host.macAddress}
            </div>
            <div className="flex items-center gap-1.5 flex-wrap">
              {authorized && (
                <Badge variant="outline" className="text-[10px] px-1.5 py-0 h-4 border-emerald-500/30 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 font-normal">
                  Authorized
                </Badge>
              )}
              {bypassed && (
                <Badge variant="outline" className="text-[10px] px-1.5 py-0 h-4 border-sky-500/30 bg-sky-500/10 text-sky-600 dark:text-sky-400 font-normal">
                  Bypassed
                </Badge>
              )}
              {!authorized && !bypassed && (
                <Badge variant="outline" className="text-[10px] px-1.5 py-0 h-4 font-normal text-muted-foreground">
                  Unauthenticated
                </Badge>
              )}
            </div>
          </div>
        </div>

        <div className="shrink-0">
          <HostsRowActions host={host} />
        </div>
      </div>

      {/* Grid: IP Address, To Address, Server */}
      <div className="grid grid-cols-2 gap-2 pt-1 border-t border-border/50 text-xs">
        <div className="space-y-0.5">
          <span className="text-[10px] uppercase tracking-wider text-muted-foreground font-medium">
            IP Address
          </span>
          <div className="flex items-center gap-1.5 font-mono text-foreground font-medium">
            <Globe className="h-3.5 w-3.5 text-primary shrink-0" />
            <span className="truncate">{ip}</span>
          </div>
        </div>

        <div className="space-y-0.5">
          <span className="text-[10px] uppercase tracking-wider text-muted-foreground font-medium">
            To Address (NAT)
          </span>
          <div className="flex items-center gap-1.5 font-mono text-muted-foreground">
            <Globe className="h-3.5 w-3.5 shrink-0" />
            <span className="truncate">{toIp}</span>
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
