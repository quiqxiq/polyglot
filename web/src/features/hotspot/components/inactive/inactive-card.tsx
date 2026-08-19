import { Badge } from '@/components/ui/badge'
import type { HotspotUser } from '@/gen/v1/hotspot_pb'
import { Clock, Hourglass, Server, UserX } from 'lucide-react'
import { InactiveRowActions } from './inactive-row-actions'

interface InactiveCardProps {
  user: HotspotUser
}

function formatBytes(bytesStr: string): string {
  const bytes = Number(bytesStr || 0)
  if (isNaN(bytes) || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`
}

export function InactiveCard({ user }: InactiveCardProps) {
  const isVoucher = user.name === user.password
  const profile = user.profile || 'default'
  const limitTime = user.limitUptime
  const limitBytes = user.limitBytes
  const server = user.server || 'all'

  return (
    <div
      className={`rounded-lg border bg-card p-3.5 shadow-sm space-y-2.5 transition-colors ${
        user.disabled
          ? 'opacity-60 bg-muted/40 border-muted text-muted-foreground'
          : 'hover:border-primary/40'
      }`}
    >
      {/* Header: User & Profile below username, Actions on right */}
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-2.5 min-w-0">
          <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-muted text-muted-foreground">
            <UserX className="h-4 w-4" />
          </div>
          <div className="min-w-0 space-y-0.5">
            <div className="flex items-center gap-1.5 min-w-0">
              <span className="font-mono text-sm font-semibold truncate">
                {user.name}
              </span>
              {isVoucher && (
                <Badge variant="outline" className="text-[9px] px-1 py-0 h-3.5 font-normal shrink-0">
                  vc
                </Badge>
              )}
            </div>
            <div>
              <Badge variant="outline" className="font-mono text-[10px] px-1.5 py-0 h-4 font-normal">
                {profile}
              </Badge>
            </div>
          </div>
        </div>

        <div className="shrink-0">
          <InactiveRowActions user={user} />
        </div>
      </div>

      {/* Grid: Used Uptime, Limit, Server */}
      <div className="grid grid-cols-2 gap-2 pt-1 border-t border-border/50 text-xs">
        <div className="space-y-0.5">
          <span className="text-[10px] uppercase tracking-wider text-muted-foreground font-medium">
            Used Uptime
          </span>
          <div className="flex items-center gap-1.5 font-mono text-muted-foreground">
            <Clock className="h-3.5 w-3.5 text-primary/70 shrink-0" />
            <span className="truncate">
              {user.uptime && user.uptime !== '0s' ? user.uptime : '- (Belum dipakai)'}
            </span>
          </div>
        </div>

        <div className="space-y-0.5">
          <span className="text-[10px] uppercase tracking-wider text-muted-foreground font-medium">
            Limit (Time/Quota)
          </span>
          <div className="flex items-center gap-1.5 font-mono text-muted-foreground">
            <Hourglass className="h-3.5 w-3.5 shrink-0" />
            <span className="truncate">
              {limitTime || (limitBytes ? formatBytes(limitBytes) : 'Unlimited')}
            </span>
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
