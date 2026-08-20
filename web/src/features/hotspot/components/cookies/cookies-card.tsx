import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import type { HotspotCookie } from '@/gen/v1/hotspot_pb'
import { Cookie, Hourglass, Network, Trash2 } from 'lucide-react'
import { useHotspot } from '../../context/hotspot-context'

interface CookiesCardProps {
  cookie: HotspotCookie
}

export function CookiesCard({ cookie }: CookiesCardProps) {
  const { setOpen, setCurrentCookie } = useHotspot()
  const user = cookie.user || '-'
  const mac = cookie.macAddress || '-'
  const expiresIn = cookie.expiresIn || '-'
  const domain = cookie.domain || 'hotspot'

  const handleDelete = () => {
    setCurrentCookie(cookie)
    setOpen('cookie-delete')
  }

  return (
    <div className="rounded-lg border bg-card p-3.5 shadow-sm space-y-2.5 transition-colors hover:border-primary/40">
      {/* Header: User & Domain badge below name, Delete Action on right */}
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-2.5 min-w-0">
          <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-amber-500/10 text-amber-600 dark:text-amber-400">
            <Cookie className="h-4 w-4" />
          </div>
          <div className="min-w-0 space-y-0.5">
            <div className="font-mono text-sm font-semibold truncate">
              {user}
            </div>
            <div>
              <Badge variant="outline" className="font-mono text-[10px] px-1.5 py-0 h-4 font-normal">
                {domain}
              </Badge>
            </div>
          </div>
        </div>

        <div className="shrink-0">
          <Button
            variant="ghost"
            size="icon"
            onClick={handleDelete}
            className="h-8 w-8 text-destructive hover:text-destructive hover:bg-destructive/10"
            title="Remove cookie"
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* Grid: MAC & Expires In */}
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
            Expires In
          </span>
          <div className="flex items-center gap-1.5 font-mono text-muted-foreground">
            <Hourglass className="h-3.5 w-3.5 text-primary/70 shrink-0" />
            <span className="truncate">{expiresIn}</span>
          </div>
        </div>
      </div>
    </div>
  )
}
