import { useEffect } from 'react'
import { ChevronsUpDown, Network, Plus, Server } from 'lucide-react'
import { useNavigate } from '@tanstack/react-router'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from '@/components/ui/sidebar'
import { useDevicesQuery } from '@/features/devices/api/use-devices'
import { useDeviceStore } from '@/stores/device-store'
import { useAuthStore } from '@/stores/auth-store'

export function DeviceSwitcher() {
  const { isMobile } = useSidebar()
  const navigate = useNavigate()
  const { data: devices = [], isLoading, isSuccess } = useDevicesQuery()
  const { selectedDeviceId, setSelectedDeviceId, hasHydrated } = useDeviceStore()

  const isOwner = Boolean(useAuthStore((s) => s.auth.user?.role?.includes('owner')))

  // Kelola default router:
  // 1. Jika pertama kali buka aplikasi (belum ada router tersimpan): pilih devices[0].id.
  // 2. Jika sudah pernah memilih router: pertahankan router tersebut seterusnya.
  // 3. Hanya fallback ke devices[0].id jika router tersimpan sudah tidak ada di inventaris.
  useEffect(() => {
    if (isLoading || !hasHydrated) return

    if (devices.length > 0) {
      const exists = devices.some((d) => d.id === selectedDeviceId)
      if (!selectedDeviceId || !exists) {
        setSelectedDeviceId(devices[0].id)
      }
    } else if (isSuccess && devices.length === 0 && selectedDeviceId) {
      setSelectedDeviceId('')
    }
  }, [devices, isLoading, isSuccess, hasHydrated, selectedDeviceId, setSelectedDeviceId])

  const currentDevice = devices.find((d) => d.id === selectedDeviceId)

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <SidebarMenuButton
              size='lg'
              className='data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground'
            >
              <div className='flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground'>
                <Server className='size-4' />
              </div>
              <div className='grid flex-1 text-start text-sm leading-tight'>
                <span className='truncate font-semibold'>
                  {isLoading
                    ? 'Loading routers...'
                    : currentDevice
                      ? currentDevice.name
                      : devices.length === 0
                        ? 'No Assigned Router'
                        : 'No Router Selected'}
                </span>
                <span className='truncate text-xs text-muted-foreground'>
                  {currentDevice
                    ? `${currentDevice.host} (${currentDevice.vendor})`
                    : devices.length === 0
                      ? 'No access assigned'
                      : 'Select MikroTik'}
                </span>
              </div>
              <ChevronsUpDown className='ms-auto' />
            </SidebarMenuButton>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            className='w-(--radix-dropdown-menu-trigger-width) min-w-64 rounded-lg'
            align='start'
            side={isMobile ? 'bottom' : 'right'}
            sideOffset={4}
          >
            <DropdownMenuLabel className='text-xs text-muted-foreground'>
              Active Routers / Sessions
            </DropdownMenuLabel>
            {devices.length === 0 ? (
              <div className='p-3 text-xs text-muted-foreground text-center space-y-1'>
                <p className='font-medium text-foreground'>No Assigned Routers</p>
                <p className='text-[11px] leading-snug'>
                  {isOwner
                    ? 'No routers found in inventory. Add one in Devices.'
                    : 'You do not have any assigned MikroTik routers. Contact Owner or Admin.'}
                </p>
              </div>
            ) : (
              devices.map((device) => (
                <DropdownMenuItem
                  key={device.id}
                  onClick={() => setSelectedDeviceId(device.id)}
                  className={`gap-2 p-2 cursor-pointer ${
                    device.id === selectedDeviceId ? 'bg-accent font-medium' : ''
                  }`}
                >
                  <div className='flex size-6 items-center justify-center rounded-sm border bg-background'>
                    <Network className='size-3.5 shrink-0' />
                  </div>
                  <div className='flex flex-col flex-1'>
                    <span className='text-sm leading-none'>{device.name}</span>
                    <span className='text-xs text-muted-foreground mt-0.5'>
                      {device.host} • {device.vendor}
                    </span>
                  </div>
                </DropdownMenuItem>
              ))
            )}
            {isOwner && (
              <>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  className='gap-2 p-2 cursor-pointer'
                  onClick={() => navigate({ to: '/devices' })}
                >
                  <div className='flex size-6 items-center justify-center rounded-md border bg-background'>
                    <Plus className='size-4' />
                  </div>
                  <div className='font-medium text-muted-foreground'>Manage Devices</div>
                </DropdownMenuItem>
              </>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  )
}
