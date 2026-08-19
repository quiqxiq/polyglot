import { Badge } from '@/components/ui/badge'
import type { PPPProfile } from '@/gen/v1/ppp_pb'
import { Globe, Layers, Server, Users } from 'lucide-react'
import { ProfilesRowActions } from './profiles-row-actions'

interface ProfilesCardProps {
  profile: PPPProfile
}

export function ProfilesCard({ profile }: ProfilesCardProps) {
  const rateLimit = profile.rateLimit || 'Unlimited'
  const remote = profile.remoteAddress || '-'
  const local = profile.localAddress || '-'
  const dns = profile.dnsServer || '-'
  const shared = profile.sharedUsers || '1'
  const onlyOne = profile.onlyOne || 'default'

  return (
    <div className="rounded-lg border bg-card p-3.5 shadow-sm space-y-2.5 transition-colors hover:border-primary/40">
      {/* Header: Profile Name & Rate Limit below name, Actions on right */}
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-2.5 min-w-0">
          <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-primary/10 text-primary">
            <Layers className="h-4 w-4" />
          </div>
          <div className="min-w-0 space-y-0.5">
            <div className="font-mono text-sm font-semibold truncate">
              {profile.name}
            </div>
            <div>
              <Badge variant="outline" className="font-mono text-[10px] px-1.5 py-0 h-4 font-normal">
                {rateLimit}
              </Badge>
            </div>
          </div>
        </div>

        <div className="shrink-0">
          <ProfilesRowActions row={profile} />
        </div>
      </div>

      {/* Grid: Pools, DNS, Shared */}
      <div className="grid grid-cols-2 gap-2 pt-1 border-t border-border/50 text-xs">
        <div className="space-y-0.5">
          <span className="text-[10px] uppercase tracking-wider text-muted-foreground font-medium">
            Remote Pool / IP
          </span>
          <div className="flex items-center gap-1.5 font-mono text-muted-foreground">
            <Server className="h-3.5 w-3.5 text-primary/70 shrink-0" />
            <span className="truncate">{remote}</span>
          </div>
        </div>

        <div className="space-y-0.5">
          <span className="text-[10px] uppercase tracking-wider text-muted-foreground font-medium">
            Local Pool / IP
          </span>
          <div className="flex items-center gap-1.5 font-mono text-muted-foreground">
            <Server className="h-3.5 w-3.5 shrink-0" />
            <span className="truncate">{local}</span>
          </div>
        </div>

        <div className="space-y-0.5">
          <span className="text-[10px] uppercase tracking-wider text-muted-foreground font-medium">
            DNS Server
          </span>
          <div className="flex items-center gap-1.5 font-mono text-muted-foreground">
            <Globe className="h-3.5 w-3.5 shrink-0" />
            <span className="truncate">{dns}</span>
          </div>
        </div>

        <div className="space-y-0.5">
          <span className="text-[10px] uppercase tracking-wider text-muted-foreground font-medium">
            Shared / Only-One
          </span>
          <div className="flex items-center gap-1.5 font-mono text-muted-foreground">
            <Users className="h-3.5 w-3.5 shrink-0" />
            <span className="truncate">{shared} User • {onlyOne}</span>
          </div>
        </div>
      </div>
    </div>
  )
}
