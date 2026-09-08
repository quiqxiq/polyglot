import type { Device } from '@/gen/v1/device_pb'
import { Server, FileSpreadsheet, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

interface PlansImportSourceTabsProps {
  activeTab: 'router' | 'file'
  onTabChange: (tab: 'router' | 'file') => void
  devices: Device[]
  isDevicesLoading: boolean
  selectedDeviceId: string
  onSelectDevice: (id: string) => void
  serviceType: string
  onSelectServiceType: (st: string) => void
  onPullRouter: () => void
  isPulling: boolean
}

export function PlansImportSourceTabs({
  activeTab,
  onTabChange,
  devices,
  isDevicesLoading,
  selectedDeviceId,
  onSelectDevice,
  serviceType,
  onSelectServiceType,
  onPullRouter,
  isPulling,
}: PlansImportSourceTabsProps) {
  return (
    <Tabs
      value={activeTab}
      onValueChange={(v) => onTabChange(v as 'router' | 'file')}
      className='space-y-4'
    >
      <TabsList className='grid w-full grid-cols-2 sm:w-[480px]'>
        <TabsTrigger value='router' className='flex items-center gap-2'>
          <Server className='h-4 w-4' />
          Metode A: Tarik dari Router
        </TabsTrigger>
        <TabsTrigger value='file' className='flex items-center gap-2'>
          <FileSpreadsheet className='h-4 w-4' />
          Metode B: Upload File
        </TabsTrigger>
      </TabsList>

      {/* ─── TAB A: TARIK PROFIL ROUTER ───────────────────────────────────── */}
      <TabsContent value='router'>
        <Card>
          <CardHeader className='pb-3'>
            <CardTitle className='text-base'>
              Tarik Profil Layanan Langsung dari MikroTik
            </CardTitle>
            <CardDescription className='text-xs'>
              Membaca daftar profile dari <code>/ppp/profile</code> (PPPoE) dan{' '}
              <code>/ip/hotspot/user/profile</code> (Hotspot).
            </CardDescription>
          </CardHeader>
          <CardContent className='space-y-4'>
            <div className='flex flex-wrap items-end gap-3'>
              {/* Router Selector */}
              <div className='w-full sm:w-72'>
                <label className='mb-1 block text-xs font-medium text-muted-foreground'>
                  Pilih Router Target
                </label>
                <Select value={selectedDeviceId} onValueChange={onSelectDevice}>
                  <SelectTrigger className='h-9 w-full text-xs sm:text-sm'>
                    <SelectValue placeholder='Pilih Router MikroTik...' />
                  </SelectTrigger>
                  <SelectContent>
                    {isDevicesLoading ? (
                      <SelectItem disabled value='loading'>
                        Memuat data router...
                      </SelectItem>
                    ) : devices.length === 0 ? (
                      <SelectItem disabled value='empty'>
                        Tidak ada router terhubung
                      </SelectItem>
                    ) : (
                      devices.map((d) => (
                        <SelectItem key={d.id} value={d.id} className='text-xs'>
                          {d.name} ({d.host})
                        </SelectItem>
                      ))
                    )}
                  </SelectContent>
                </Select>
              </div>

              {/* Service Type Filter */}
              <div className='w-full sm:w-48'>
                <label className='mb-1 block text-xs font-medium text-muted-foreground'>
                  Tipe Layanan
                </label>
                <Select value={serviceType} onValueChange={onSelectServiceType}>
                  <SelectTrigger className='h-9 w-full text-xs sm:text-sm'>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value='ALL' className='text-xs'>
                      Semua (PPPoE + Hotspot)
                    </SelectItem>
                    <SelectItem value='PPPOE' className='text-xs'>
                      Hanya PPPoE
                    </SelectItem>
                    <SelectItem value='HOTSPOT' className='text-xs'>
                      Hanya Hotspot
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {/* Action Button */}
              <Button
                type='button'
                onClick={onPullRouter}
                disabled={!selectedDeviceId || isPulling}
                className='h-9'
              >
                {isPulling ? (
                  <>
                    <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                    <span>Menarik Profil...</span>
                  </>
                ) : (
                  <>
                    <Server className='mr-2 h-4 w-4' />
                    <span>Tarik Profil Router</span>
                  </>
                )}
              </Button>
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      {/* ─── TAB B: UPLOAD FILE ────────────────────────────────────────────── */}
      <TabsContent value='file'>
        <Card>
          <CardHeader className='pb-3'>
            <CardTitle className='text-base'>
              Import Profil Paket dari Spreadsheet
            </CardTitle>
            <CardDescription className='text-xs'>
              Mendukung file XLSX atau CSV hasil ekspor billing sebelumnya.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className='flex flex-col items-center justify-center rounded-lg border-2 border-dashed p-8 text-center'>
              <FileSpreadsheet className='h-10 w-10 text-muted-foreground/60' />
              <p className='mt-2 text-sm font-medium'>
                Fitur Unggah File Service Plans
              </p>
              <p className='mt-1 text-xs text-muted-foreground'>
                Gunakan Metode A (Tarik dari Router) untuk sinkronisasi otomatis
                langsung dari MikroTik.
              </p>
            </div>
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>
  )
}
