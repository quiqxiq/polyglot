import { Badge } from '@/components/ui/badge'
import type { PPPSecret } from '@/gen/v1/ppp_pb'
import { Clock, Globe, Server, User } from 'lucide-react'
import { SecretsRowActions } from './secrets-row-actions'

interface SecretsCardProps {
  secret: PPPSecret
}

export function SecretsCard({ secret }: SecretsCardProps) {
  const profile = secret.profile || 'default'
  const service = (secret.service || 'any').toUpperCase()
  const remoteIp = secret.remoteAddress || '-'
  const lastLogout = secret.lastLoggedOut || 'Never'

  return (
    <div
      className={`rounded-lg border bg-card p-3.5 shadow-sm space-y-2.5 transition-colors ${
        secret.disabled
          ? 'opacity-60 bg-muted/40 border-muted text-muted-foreground'
          : 'hover:border-primary/40'
      }`}
    >
      {/* Header: User & Profile below username, Actions on right */}
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-2.5 min-w-0">
          <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-muted text-muted-foreground">
            <User className="h-4 w-4" />
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
          <SecretsRowActions row={secret} />
        </div>
      </div>

      {/* Grid: Remote IP, Service, Last Logout */}
      <div className="grid grid-cols-2 gap-2 pt-1 border-t border-border/50 text-xs">
        <div className="space-y-0.5">
          <span className="text-[10px] uppercase tracking-wider text-muted-foreground font-medium">
            Remote IP
          </span>
          <div className="flex items-center gap-1.5 font-mono text-muted-foreground">
            <Globe className="h-3.5 w-3.5 text-primary/70 shrink-0" />
            <span className="truncate">{remoteIp}</span>
          </div>
        </div>

        <div className="space-y-0.5">
          <span className="text-[10px] uppercase tracking-wider text-muted-foreground font-medium">
            Service
          </span>
          <div className="flex items-center gap-1.5 font-mono text-muted-foreground">
            <Server className="h-3.5 w-3.5 shrink-0" />
            <span className="truncate">{service}</span>
          </div>
        </div>

        <div className="col-span-2 space-y-0.5">
          <span className="text-[10px] uppercase tracking-wider text-muted-foreground font-medium">
            Last Logout
          </span>
          <div className="flex items-center gap-1.5 font-mono text-muted-foreground">
            <Clock className="h-3.5 w-3.5 text-primary/70 shrink-0" />
            <span className="truncate">{lastLogout}</span>
          </div>
        </div>
      </div>
    </div>
  )
}
