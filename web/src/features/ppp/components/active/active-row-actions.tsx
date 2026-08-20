import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import type { PPPActiveSession } from '@/gen/v1/ppp_pb'
import { Activity, ExternalLink, Unplug } from 'lucide-react'
import { usePPP } from '../../context/ppp-context'

interface ActiveRowActionsProps {
  row: PPPActiveSession
}

export function ActiveRowActions({ row }: ActiveRowActionsProps) {
  const { setOpen, setCurrentActiveSession } = usePPP()

  const handleKick = () => {
    setCurrentActiveSession(row)
    setOpen('active-kick')
  }

  const handlePing = () => {
    setCurrentActiveSession(row)
    setOpen('active-ping')
  }

  const handleOpenIp = () => {
    if (!row.address) return
    window.open(`http://${row.address}`, '_blank', 'noopener,noreferrer')
  }

  return (
    <div className="flex items-center justify-end gap-1">
      {/* 1. Open IP Button */}
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7 text-muted-foreground hover:bg-primary/10 hover:text-primary"
            onClick={handleOpenIp}
            disabled={!row.address}
          >
            <ExternalLink className="h-3.5 w-3.5" />
            <span className="sr-only">Open IP</span>
          </Button>
        </TooltipTrigger>
        <TooltipContent side="top">
          {row.address ? `Open http://${row.address} in new tab` : 'No IP assigned'}
        </TooltipContent>
      </Tooltip>

      {/* 2. Live Ping Button */}
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7 text-muted-foreground hover:bg-emerald-500/10 hover:text-emerald-600 dark:hover:text-emerald-400"
            onClick={handlePing}
            disabled={!row.address}
          >
            <Activity className="h-3.5 w-3.5" />
            <span className="sr-only">Live Ping</span>
          </Button>
        </TooltipTrigger>
        <TooltipContent side="top">
          {row.address ? `Stream Ping to ${row.address}` : 'No IP to ping'}
        </TooltipContent>
      </Tooltip>

      {/* 3. Disconnect Button */}
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7 text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
            onClick={handleKick}
          >
            <Unplug className="h-3.5 w-3.5" />
            <span className="sr-only">Disconnect</span>
          </Button>
        </TooltipTrigger>
        <TooltipContent side="top">Disconnect active session</TooltipContent>
      </Tooltip>
    </div>
  )
}
