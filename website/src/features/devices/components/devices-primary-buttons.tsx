import { PlusIcon, GridIcon, TableIcon } from '@radix-ui/react-icons'
import { Button } from '@/components/ui/button'
import { useDevicesContext } from './devices-provider'

export function DevicesPrimaryButtons() {
  const { setOpen, setCurrentRow, viewMode, setViewMode } = useDevicesContext()

  return (
    <div className='flex items-center gap-2'>
      {/* View Mode Switcher (Grid Cards vs Table) */}
      <div className='flex items-center rounded-lg border bg-muted p-1 gap-1'>
        <Button
          variant={viewMode === 'card' ? 'secondary' : 'ghost'}
          size='sm'
          className='h-7 px-2 text-xs gap-1.5'
          onClick={() => setViewMode('card')}
          title='Card Grid View'
        >
          <GridIcon className='h-3.5 w-3.5' />
          <span className='hidden sm:inline'>Cards</span>
        </Button>
        <Button
          variant={viewMode === 'table' ? 'secondary' : 'ghost'}
          size='sm'
          className='h-7 px-2 text-xs gap-1.5'
          onClick={() => setViewMode('table')}
          title='Table View'
        >
          <TableIcon className='h-3.5 w-3.5' />
          <span className='hidden sm:inline'>Table</span>
        </Button>
      </div>

      <Button
        onClick={() => {
          setCurrentRow(null)
          setOpen('add')
        }}
        className='gap-1.5 h-9 text-xs sm:text-sm'
      >
        <PlusIcon className='h-4 w-4' />
        Add Device
      </Button>
    </div>
  )
}
