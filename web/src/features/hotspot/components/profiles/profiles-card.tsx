import { Badge } from '@/components/ui/badge'
import type { HotspotProfile } from '@/gen/v1/hotspot_pb'
import { Banknote, Gauge, Hourglass, Layers } from 'lucide-react'
import { ProfilesRowActions } from './profiles-row-actions'

interface ProfilesCardProps {
  profile: HotspotProfile
}

export function ProfilesCard({ profile }: ProfilesCardProps) {
  const rateLimit = profile.rateLimit || 'Unlimited'
  const shared = profile.sharedUsers || '1'
  const validity = profile.validity || '-'
  const mode = profile.modeExpire || 'None'
  const sprice = Number(profile.sellingPrice || 0)
  const price = Number(profile.price || 0)

  return (
    <div className="rounded-lg border bg-card p-3.5 shadow-sm space-y-2.5 transition-colors hover:border-primary/40">
      {/* Header: Profile Name & Shared Badge below name, Actions on right */}
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
                Shared: {shared} User
              </Badge>
            </div>
          </div>
        </div>

        <div className="shrink-0">
          <ProfilesRowActions profile={profile} />
        </div>
      </div>

      {/* Grid: Rate Limit, Validity/Mode, Prices */}
      <div className="grid grid-cols-2 gap-2 pt-1 border-t border-border/50 text-xs">
        <div className="space-y-0.5">
          <span className="text-[10px] uppercase tracking-wider text-muted-foreground font-medium">
            Rate Limit
          </span>
          <div className="flex items-center gap-1.5 font-mono text-muted-foreground">
            <Gauge className="h-3.5 w-3.5 text-primary/70 shrink-0" />
            <span className="truncate">{rateLimit}</span>
          </div>
        </div>

        <div className="space-y-0.5">
          <span className="text-[10px] uppercase tracking-wider text-muted-foreground font-medium">
            Validity / Mode
          </span>
          <div className="flex items-center gap-1.5 font-mono text-muted-foreground">
            <Hourglass className="h-3.5 w-3.5 shrink-0" />
            <span className="truncate">{validity} • {mode}</span>
          </div>
        </div>

        <div className="col-span-2 space-y-0.5">
          <span className="text-[10px] uppercase tracking-wider text-muted-foreground font-medium">
            Harga Jual / Beli
          </span>
          <div className="flex items-center gap-1.5 font-mono">
            <Banknote className="h-3.5 w-3.5 text-emerald-500 shrink-0" />
            <span className="font-semibold text-emerald-600 dark:text-emerald-400">
              {sprice > 0 ? `Rp ${sprice.toLocaleString('id-ID')}` : '-'}
            </span>
            {price > 0 && (
              <span className="text-muted-foreground text-[11px]">
                (Modal: Rp {price.toLocaleString('id-ID')})
              </span>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
