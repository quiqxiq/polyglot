import { Badge } from '@/components/ui/badge'
import type { PPPSecret } from '@/gen/v1/ppp_pb'
import { Clock, Network, UserX } from 'lucide-react'
import { InactiveRowActions } from './inactive-row-actions'

interface InactiveCardProps {
  secret: PPPSecret
}

export function InactiveCard({ secret }: InactiveCardProps) {
  const profile = secret.profile || 'default'
  const lastLogout = secret.lastLoggedOut || 'Never'
  const mac = secret.callerId || '-'

  return (
    <div
      className={`rounded-lg border bg-card p-3.5 shadow-sm space-y-2.5 transition-colors ${
        secret.disabled
          ? 'opacity-60 bg-muted/40 border-muted text-muted-foreground'
          : 'hover:border-primary/40'
      }`}
    >
      {/* Header: User & Profile below username, Action on right */}
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-2.5 min-w-0">
          <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-muted text-muted-foreground">
            <UserX className="h-4 w-4" />
          </div>
          <div className="min-w-0 space-y-0.5">
            <div className="font-mono text-sm font-semibold truncate">
              {secret.name}
            </div>
            <div>
              <Badge variant="outline" className="font-mono text-[10px] px-1.5 py-0 h-4 font-normal">
                {profile}
              </Badge>
            </div>
          </div>
        </div>

        <div className="shrink-0">
          <InactiveRowActions row={secret} />
        </div>
      </div>

      {/* Grid: Last Logout & MAC Address */}
      <div className="grid grid-cols-2 gap-2 pt-1 border-t border-border/50 text-xs">
        <div className="space-y-0.5">
          <span className="text-[10px] uppercase tracking-wider text-muted-foreground font-medium">
            Last Logout
          </span>
          <div className="flex items-center gap-1.5 font-mono text-muted-foreground">
            <Clock className="h-3.5 w-3.5 text-primary/70 shrink-0" />
            <span className="truncate">{lastLogout}</span>
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
      </div>
    </div>
  )
}
