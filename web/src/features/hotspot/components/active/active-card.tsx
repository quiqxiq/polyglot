import { Badge } from '@/components/ui/badge'
import type { HotspotActiveSession } from '@/gen/v1/hotspot_pb'
import { ArrowDown, ArrowUp, Clock, Globe, Network, UserCheck } from 'lucide-react'
import { ActiveRowActions } from './active-row-actions'

interface ActiveCardProps {
  session: HotspotActiveSession
}

function formatBytes(bytesStr: string): string {
  const bytes = Number(bytesStr || 0)
  if (isNaN(bytes) || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`
}

export function ActiveCard({ session }: ActiveCardProps) {
  const server = session.server || 'all'
  const ip = session.address || '-'
  const uptime = session.uptime || '-'
  const mac = session.macAddress || '-'

  return (
    <div className="rounded-lg border bg-card p-3.5 shadow-sm space-y-2.5 transition-colors hover:border-primary/40">
      {/* Header: User with pulse indicator, Server badge below name, Actions on right */}
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-2.5 min-w-0">
          <div className="relative flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
            <UserCheck className="h-4 w-4" />
            <span className="absolute -top-0.5 -right-0.5 flex h-2 w-2">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
              <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500" />
            </span>
          </div>
          <div className="min-w-0 space-y-0.5">
            <div className="font-mono text-sm font-semibold truncate">
              {session.user}
            </div>
            <div>
              <Badge variant="outline" className="font-mono text-[10px] px-1.5 py-0 h-4 font-normal">
                {server}
              </Badge>
            </div>
          </div>
        </div>

        <div className="shrink-0">
          <ActiveRowActions session={session} />
        </div>
      </div>

      {/* Grid: IP Address, Uptime, MAC, Traffic */}
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
            Uptime
          </span>
          <div className="flex items-center gap-1.5 font-mono text-muted-foreground">
            <Clock className="h-3.5 w-3.5 text-emerald-500 shrink-0" />
            <span className="truncate">{uptime}</span>
          </div>
        </div>

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
            Traffic (In / Out)
          </span>
          <div className="flex items-center gap-2 font-mono text-muted-foreground">
            <span className="flex items-center gap-0.5 text-emerald-600 dark:text-emerald-400">
              <ArrowDown className="h-3 w-3" />
              {formatBytes(session.bytesOut)}
            </span>
            <span className="flex items-center gap-0.5 text-sky-600 dark:text-sky-400">
              <ArrowUp className="h-3 w-3" />
              {formatBytes(session.bytesIn)}
            </span>
          </div>
        </div>
      </div>
    </div>
  )
}
