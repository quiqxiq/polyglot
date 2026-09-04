import { Badge } from '@/components/ui/badge'
import type { EnrichedPPPActiveSession } from '../../api/use-ppp-stream'
import { formatUptime } from '../../utils/format-uptime'
import { Clock, Globe, Network, UserCheck } from 'lucide-react'
import { ActiveRowActions } from './active-row-actions'

interface ActiveCardProps {
  session: EnrichedPPPActiveSession
}

export function ActiveCard({ session }: ActiveCardProps) {
  const profile = session.profile || 'default'
  const ip = session.address || '-'
  const uptime = formatUptime(session.uptime)
  const mac = session.callerId || '-'
  const service = session.service || 'pppoe'

  return (
    <div className="rounded-lg border bg-card p-3.5 shadow-sm space-y-2.5 transition-colors hover:border-primary/40">
      {/* Header: User with online indicator, Profile & Service badges below name, and Actions on right */}
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
              {session.name}
            </div>
            <div className="flex items-center gap-1.5">
              <Badge variant="outline" className="font-mono text-[10px] px-1.5 py-0 h-4 font-normal">
                {profile}
              </Badge>
              <Badge variant="secondary" className="font-mono text-[10px] px-1.5 py-0 h-4 font-normal uppercase">
                {service}
              </Badge>
            </div>
          </div>
        </div>

        <div className="shrink-0">
          <ActiveRowActions row={session} />
        </div>
      </div>

      {/* Grid: IP Address, Uptime, MAC Address */}
      <div className="grid grid-cols-2 gap-2 pt-1 border-t border-border/50 text-xs">
        <div className="space-y-0.5">
          <span className="text-[10px] uppercase tracking-wider text-muted-foreground font-medium">
            IP Address
          </span>
          <div className="flex items-center gap-1.5 font-mono font-medium text-foreground">
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

        <div className="col-span-2 space-y-0.5">
          <span className="text-[10px] uppercase tracking-wider text-muted-foreground font-medium">
            MAC Address / Caller ID
          </span>
          <div className="flex items-center gap-1.5 font-mono text-muted-foreground">
            <Network className="h-3.5 w-3.5 shrink-0" />
            <span className="truncate">{mac}</span>
          </div>
        </div>
      </div>
    </div>
  )
}
