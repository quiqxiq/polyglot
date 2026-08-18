import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import type { PPPActiveSession } from '@/gen/v1/ppp_pb'
import { Unplug } from 'lucide-react'
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

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button
          variant="ghost"
          size="sm"
          className="h-8 px-2 text-destructive hover:bg-destructive/10 hover:text-destructive"
          onClick={handleKick}
        >
          <Unplug className="mr-1.5 h-3.5 w-3.5" />
          Disconnect
        </Button>
      </TooltipTrigger>
      <TooltipContent>Kick & disconnect this active PPPoE session</TooltipContent>
    </Tooltip>
  )
}
