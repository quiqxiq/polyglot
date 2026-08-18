import { Cookie, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { CookiesTable } from './cookies-table'
import { useHotspotCookiesQuery } from '../../api/use-hotspot-cookies'
import { useDeviceStore } from '@/stores/device-store'
import { useHotspot } from '../../context/hotspot-context'

export function CookiesTab() {
  const { selectedDeviceId } = useDeviceStore()
  const { setOpen } = useHotspot()
  const { data: cookies = [], isLoading } = useHotspotCookiesQuery(selectedDeviceId || '')

  return (
    <div className='flex flex-1 flex-col gap-4'>
      <div className='flex items-center justify-between'>
        <div className='flex items-center gap-2'>
          <Cookie className='size-5 text-amber-500' />
          <div>
            <h3 className='text-sm font-semibold'>Hotspot Cookies</h3>
            <p className='text-xs text-muted-foreground'>
              Active MAC login cookies stored in the router for seamless reconnects.
            </p>
          </div>
        </div>
        <Button
          variant='destructive'
          size='sm'
          onClick={() => setOpen('cookie-clear-all')}
          disabled={cookies.length === 0 || isLoading}
          className='gap-1.5 h-8'
        >
          <Trash2 className='size-3.5' />
          Clear All Cookies
        </Button>
      </div>

      <CookiesTable data={cookies} isLoading={isLoading} />
    </div>
  )
}
