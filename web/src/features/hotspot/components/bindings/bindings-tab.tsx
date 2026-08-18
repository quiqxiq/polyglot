import { Plus, ShieldCheck } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { BindingsTable } from './bindings-table'
import { useHotspotIPBindingsQuery } from '../../api/use-hotspot-bindings'
import { useDeviceStore } from '@/stores/device-store'
import { useHotspot } from '../../context/hotspot-context'

export function BindingsTab() {
  const { selectedDeviceId } = useDeviceStore()
  const { setOpen, setCurrentBinding } = useHotspot()
  const { data: bindings = [], isLoading } = useHotspotIPBindingsQuery(selectedDeviceId || '')

  return (
    <div className='flex flex-1 flex-col gap-4'>
      <div className='flex items-center justify-between'>
        <div className='flex items-center gap-2'>
          <ShieldCheck className='size-5 text-emerald-500' />
          <div>
            <h3 className='text-sm font-semibold'>Hotspot IP Bindings</h3>
            <p className='text-xs text-muted-foreground'>
              Whitelist, bypass, or block devices without entering hotspot login credentials.
            </p>
          </div>
        </div>
        <Button
          size='sm'
          onClick={() => {
            setCurrentBinding(null)
            setOpen('binding-create')
          }}
          className='gap-1.5 h-8'
        >
          <Plus className='size-3.5' />
          Add Binding
        </Button>
      </div>

      <BindingsTable data={bindings} isLoading={isLoading} />
    </div>
  )
}
