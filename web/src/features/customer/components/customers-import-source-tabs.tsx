import type { Device } from '@/gen/v1/device_pb'
import { Server, FileSpreadsheet, Loader2, Info } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

interface CustomersImportSourceTabsProps {
  activeTab: 'router' | 'file'
  onTabChange: (tab: 'router' | 'file') => void
  devices: Device[]
  isDevicesLoading: boolean
  selectedDeviceId: string
  onSelectDevice: (id: string) => void
  pullPPPoE: boolean
  onTogglePPPoE: (v: boolean) => void
  pullHotspotMember: boolean
  onToggleHotspotMember: (v: boolean) => void
  pullHotspotIPBinding: boolean
  onToggleHotspotIPBinding: (v: boolean) => void
  pullHotspotVoucher: boolean
  onToggleHotspotVoucher: (v: boolean) => void
  onPullRouter: () => void
  isPulling: boolean
}

export function CustomersImportSourceTabs({
  activeTab,
  onTabChange,
  devices,
  isDevicesLoading,
  selectedDeviceId,
  onSelectDevice,
  pullPPPoE,
  onTogglePPPoE,
  pullHotspotMember,
  onToggleHotspotMember,
  pullHotspotIPBinding,
  onToggleHotspotIPBinding,
  pullHotspotVoucher,
  onToggleHotspotVoucher,
  onPullRouter,
  isPulling,
}: CustomersImportSourceTabsProps) {
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

      {/* TAB METODE A: Tarik dari Router */}
      <TabsContent value='router' className='space-y-4'>
        <Card className='border-border/60'>
          <CardHeader className='pb-3'>
            <CardTitle className='text-base font-semibold'>
              Pilih Router Target & Sumber Akun
            </CardTitle>
          </CardHeader>
          <CardContent className='space-y-4'>
            <div className='grid grid-cols-1 items-end gap-4 sm:grid-cols-3'>
              <div className='space-y-1.5'>
                <label className='text-xs font-medium text-muted-foreground'>
                  Router MikroTik
                </label>
                <Select
                  value={selectedDeviceId}
                  onValueChange={onSelectDevice}
                  disabled={isDevicesLoading}
                >
                  <SelectTrigger>
                    <SelectValue placeholder='Pilih Router...' />
                  </SelectTrigger>
                  <SelectContent>
                    {devices.map((d) => (
                      <SelectItem key={d.id} value={d.id}>
                        <span className='font-medium'>{d.name}</span>
                        <span className='ml-2 text-xs text-muted-foreground'>
                          ({d.host})
                        </span>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {/* Checkbox Sumber Akun */}
              <div className='flex flex-wrap items-center gap-4 py-1 sm:col-span-2'>
                <label className='flex cursor-pointer items-center gap-2 text-xs font-medium'>
                  <Checkbox
                    checked={pullPPPoE}
                    onCheckedChange={(c) => onTogglePPPoE(Boolean(c))}
                  />
                  <span>PPPoE Secrets</span>
                </label>

                <label className='flex cursor-pointer items-center gap-2 text-xs font-medium'>
                  <Checkbox
                    checked={pullHotspotMember}
                    onCheckedChange={(c) => onToggleHotspotMember(Boolean(c))}
                  />
                  <span>Hotspot Members</span>
                </label>

                <label className='flex cursor-pointer items-center gap-2 text-xs font-medium'>
                  <Checkbox
                    checked={pullHotspotIPBinding}
                    onCheckedChange={(c) =>
                      onToggleHotspotIPBinding(Boolean(c))
                    }
                  />
                  <span>IP Bindings</span>
                </label>

                <label className='flex cursor-pointer items-center gap-2 text-xs font-medium text-muted-foreground'>
                  <Checkbox
                    checked={pullHotspotVoucher}
                    onCheckedChange={(c) => onToggleHotspotVoucher(Boolean(c))}
                  />
                  <span>Hotspot Vouchers</span>
                </label>
              </div>
            </div>

            <div className='flex items-center justify-between border-t pt-2'>
              <div className='flex items-center gap-1.5 text-xs text-muted-foreground'>
                <Info className='h-3.5 w-3.5 text-sky-500' />
                <span>
                  Secara default, voucher jangka pendek dilewati agar tidak
                  mengotori data pelanggan bulanan.
                </span>
              </div>

              <Button
                onClick={onPullRouter}
                disabled={!selectedDeviceId || isPulling}
                className='space-x-2'
              >
                {isPulling ? (
                  <>
                    <Loader2 className='h-4 w-4 animate-spin' />
                    <span>Membaca Data Router...</span>
                  </>
                ) : (
                  <>
                    <Server className='h-4 w-4' />
                    <span>Tarik Akun dari Router</span>
                  </>
                )}
              </Button>
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      {/* TAB METODE B: Upload File */}
      <TabsContent value='file' className='space-y-4'>
        <Card className='border-border/60'>
          <CardHeader className='pb-3'>
            <CardTitle className='text-base font-semibold'>
              Upload File Spreadsheet Pelanggan
            </CardTitle>
          </CardHeader>
          <CardContent className='space-y-4'>
            <div className='rounded-lg border-2 border-dashed p-6 text-center text-muted-foreground'>
              <FileSpreadsheet className='mx-auto mb-2 h-8 w-8 opacity-50' />
              <p className='text-sm font-medium'>
                Format CSV atau Excel untuk Pelanggan & Langganan
              </p>
              <p className='mt-1 text-xs text-muted-foreground'>
                Kolom: Username, Password, Name, Phone, Address, ServiceType,
                Plan, Price
              </p>
              <Input
                type='file'
                accept='.csv,.xlsx,.xls'
                className='mx-auto mt-4 max-w-xs cursor-pointer text-xs'
                onChange={() => {
                  toast.info(
                    'File terpilih. Untuk sinkronisasi otomatis, gunakan Metode A Tarik dari Router.'
                  )
                }}
              />
            </div>
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>
  )
}
