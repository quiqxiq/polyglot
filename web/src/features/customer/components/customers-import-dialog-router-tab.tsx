import { useState } from 'react'
import type { Device } from '@/gen/v1/device_pb'
import { ImportRouterRequest } from '@/gen/v1/ispadmin_pb'
import {
  CheckCircle2,
  Eye,
  Loader2,
  Info,
  ShieldCheck,
} from 'lucide-react'
import { toast } from 'sonner'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { useImportRouterMutation } from '../api/use-customer'

interface CustomersImportDialogRouterTabProps {
  devices: Device[]
  devicesLoading: boolean
  onSuccess: () => void
}

export function CustomersImportDialogRouterTab({
  devices,
  devicesLoading,
  onSuccess,
}: CustomersImportDialogRouterTabProps) {
  const importRouter = useImportRouterMutation()
  const [selectedDeviceId, setSelectedDeviceId] = useState<string>('')
  const [pullPPPoE, setPullPPPoE] = useState<boolean>(true)
  const [pullHotspotMember, setPullHotspotMember] = useState<boolean>(true)
  const [pullHotspotIPBinding, setPullHotspotIPBinding] =
    useState<boolean>(true)
  const [pullHotspotVoucher, setPullHotspotVoucher] = useState<boolean>(false)
  const [routerPreview, setRouterPreview] = useState<{
    rowsTotal: number
    pppoeDetected: number
    hotspotPermanentDetected: number
    hotspotIpBindingDetected: number
    vouchersSkipped: number
    previewRows: string[]
    validationErrors: string[]
  } | null>(null)
  const [isPullingPreview, setIsPullingPreview] = useState(false)
  const [isExecutingRouterImport, setIsExecutingRouterImport] = useState(false)

  const computeServiceType = () => {
    const hasPPPoE = pullPPPoE
    const hasHotspot =
      pullHotspotMember || pullHotspotIPBinding || pullHotspotVoucher
    if (hasPPPoE && hasHotspot) return 'ALL'
    if (hasPPPoE) return 'PPPOE'
    if (hasHotspot) return 'HOTSPOT'
    return 'ALL'
  }

  const handleRouterPreview = async () => {
    if (!selectedDeviceId) {
      toast.error('Pilih router target terlebih dahulu')
      return
    }
    const dev = devices.find((d) => d.id === selectedDeviceId)
    setIsPullingPreview(true)
    try {
      const res = await importRouter.mutateAsync(
        new ImportRouterRequest({
          deviceId: selectedDeviceId,
          deviceName: dev?.name || selectedDeviceId,
          dryRun: true,
          serviceType: computeServiceType(),
          includeIpBindings: pullHotspotIPBinding,
          includeVouchers: pullHotspotVoucher,
        })
      )
      setRouterPreview({
        rowsTotal: res.result?.rowsTotal || 0,
        pppoeDetected: res.pppoeDetected,
        hotspotPermanentDetected: res.hotspotPermanentDetected,
        hotspotIpBindingDetected: res.hotspotIpBindingDetected,
        vouchersSkipped: res.vouchersSkipped,
        previewRows: res.previewRows || [],
        validationErrors: res.validationErrors || [],
      })
      toast.success(
        `Ditemukan ${res.result?.rowsTotal || 0} akun aktif di router`
      )
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : 'Gagal menarik data dari router'
      )
    } finally {
      setIsPullingPreview(false)
    }
  }

  const handleRouterImportExecute = async () => {
    if (!selectedDeviceId) {
      toast.error('Pilih router target terlebih dahulu')
      return
    }
    const dev = devices.find((d) => d.id === selectedDeviceId)
    setIsExecutingRouterImport(true)
    try {
      const res = await importRouter.mutateAsync(
        new ImportRouterRequest({
          deviceId: selectedDeviceId,
          deviceName: dev?.name || selectedDeviceId,
          dryRun: false,
          serviceType: computeServiceType(),
          includeIpBindings: pullHotspotIPBinding,
          includeVouchers: pullHotspotVoucher,
        })
      )
      const r = res.result
      if (!r) throw new Error('Import gagal')
      toast.success(
        `Import Selesai: ${r.customersCreated} pelanggan baru, ${r.subscriptionsCreated} langganan terhubung ke router`
      )
      onSuccess()
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : 'Gagal menyimpan data import'
      )
    } finally {
      setIsExecutingRouterImport(false)
    }
  }

  return (
    <div className='space-y-4 pt-2'>
      <div className='flex items-start gap-2 rounded-lg border bg-muted/30 p-3 text-xs text-muted-foreground'>
        <ShieldCheck className='mt-0.5 h-4 w-4 shrink-0 text-emerald-500' />
        <div>
          <span className='font-semibold text-foreground'>
            Aman & Non-Destruktif:
          </span>{' '}
          Penarikan data bersifat{' '}
          <span className='font-mono underline'>read-only</span>. Status akun
          diset ke{' '}
          <Badge
            variant='outline'
            className='border-emerald-500 px-1 py-0 text-[10px] text-emerald-600'
          >
            PROVISION_OK
          </Badge>{' '}
          sehingga tidak memutus atau menimpa kredensial router yang sedang
          aktif di lapangan.
        </div>
      </div>

      {/* Pemilihan Router */}
      <div className='space-y-1.5'>
        <label className='text-xs font-semibold text-foreground'>
          Pilih Router Target
        </label>
        <Select
          value={selectedDeviceId}
          onValueChange={(val) => {
            setSelectedDeviceId(val)
            setRouterPreview(null)
          }}
        >
          <SelectTrigger className='w-full'>
            <SelectValue placeholder='Pilih Router MikroTik tujuan' />
          </SelectTrigger>
          <SelectContent>
            {devicesLoading ? (
              <SelectItem disabled value='loading'>
                Memuat daftar router…
              </SelectItem>
            ) : devices.length === 0 ? (
              <SelectItem disabled value='empty'>
                Tidak ada router terhubung
              </SelectItem>
            ) : (
              devices.map((d) => (
                <SelectItem key={d.id} value={d.id}>
                  {d.name} ({d.host})
                </SelectItem>
              ))
            )}
          </SelectContent>
        </Select>
      </div>

      {/* Checklist Sumber Akun */}
      <div className='space-y-2 rounded-lg border bg-background p-3'>
        <div className='flex items-center justify-between text-xs font-semibold text-foreground'>
          <span>Jenis Akun yang Ditarik</span>
          <span className='text-[11px] font-normal text-muted-foreground'>
            Klasifikasi Otomatis
          </span>
        </div>
        <div className='grid grid-cols-1 gap-2 text-xs sm:grid-cols-2'>
          <label className='flex cursor-pointer items-center gap-2 rounded p-1.5 transition-colors hover:bg-muted/50'>
            <Checkbox
              checked={pullPPPoE}
              onCheckedChange={(c) => setPullPPPoE(Boolean(c))}
            />
            <span>PPPoE Secrets (/ppp/secret)</span>
          </label>
          <label className='flex cursor-pointer items-center gap-2 rounded p-1.5 transition-colors hover:bg-muted/50'>
            <Checkbox
              checked={pullHotspotMember}
              onCheckedChange={(c) => setPullHotspotMember(Boolean(c))}
            />
            <span>Hotspot Member (Tanpa Script Expire)</span>
          </label>
          <label className='flex cursor-pointer items-center gap-2 rounded p-1.5 transition-colors hover:bg-muted/50'>
            <Checkbox
              checked={pullHotspotIPBinding}
              onCheckedChange={(c) => setPullHotspotIPBinding(Boolean(c))}
            />
            <span>IP Binding Statis (Bypass Portal)</span>
          </label>
          <label className='flex cursor-pointer items-center gap-2 rounded p-1.5 text-muted-foreground transition-colors hover:bg-muted/50'>
            <Checkbox
              checked={pullHotspotVoucher}
              onCheckedChange={(c) => setPullHotspotVoucher(Boolean(c))}
            />
            <span>Voucher Sementara (Mikhmon)</span>
          </label>
        </div>
        {!pullHotspotVoucher && (
          <p className='flex items-center gap-1 pt-1 text-[11px] text-muted-foreground'>
            <Info className='h-3 w-3 shrink-0 text-blue-500' />
            Voucher sementara berdurasi beberapa jam otomatis dilewati agar
            database tagihan bulanan tetap bersih.
          </p>
        )}
      </div>

      {/* Tombol Pratinjau Router */}
      <div className='flex justify-end'>
        <Button
          type='button'
          variant='outline'
          size='sm'
          onClick={handleRouterPreview}
          disabled={
            !selectedDeviceId || isPullingPreview || isExecutingRouterImport
          }
          className='gap-1.5'
        >
          {isPullingPreview ? (
            <Loader2 className='h-4 w-4 animate-spin' />
          ) : (
            <Eye className='h-4 w-4 text-primary' />
          )}
          <span>Tarik & Pratinjau Akun</span>
        </Button>
      </div>

      {/* Hasil Pratinjau Router */}
      {routerPreview && (
        <div className='space-y-3 rounded-lg border border-primary/20 bg-primary/5 p-3'>
          <div className='flex items-center justify-between text-xs font-semibold text-foreground'>
            <span className='flex items-center gap-1 text-emerald-600 dark:text-emerald-400'>
              <CheckCircle2 className='h-4 w-4' /> Hasil Deteksi Router:
            </span>
            <Badge variant='secondary' className='font-mono'>
              {routerPreview.rowsTotal} Akun Siap
            </Badge>
          </div>

          <div className='grid grid-cols-2 gap-2 text-xs sm:grid-cols-4'>
            <div className='rounded border bg-background p-2'>
              <span className='text-[10px] text-muted-foreground'>PPPoE</span>
              <p className='font-mono text-sm font-bold text-sky-600 dark:text-sky-400'>
                {routerPreview.pppoeDetected}
              </p>
            </div>
            <div className='rounded border bg-background p-2'>
              <span className='text-[10px] text-muted-foreground'>
                Hotspot Tetap
              </span>
              <p className='font-mono text-sm font-bold text-amber-600 dark:text-amber-400'>
                {routerPreview.hotspotPermanentDetected}
              </p>
            </div>
            <div className='rounded border bg-background p-2'>
              <span className='text-[10px] text-muted-foreground'>
                IP Binding
              </span>
              <p className='font-mono text-sm font-bold text-purple-600 dark:text-purple-400'>
                {routerPreview.hotspotIpBindingDetected}
              </p>
            </div>
            <div className='rounded border bg-background p-2'>
              <span className='text-[10px] text-muted-foreground'>
                Voucher Dilewati
              </span>
              <p className='font-mono text-sm font-bold text-muted-foreground'>
                {routerPreview.vouchersSkipped}
              </p>
            </div>
          </div>

          {routerPreview.previewRows.length > 0 && (
            <div className='space-y-1'>
              <span className='text-[11px] font-medium text-muted-foreground'>
                Sampel Akun (5 Teratas):
              </span>
              <ul className='space-y-1 font-mono text-[11px]'>
                {routerPreview.previewRows.slice(0, 5).map((row, i) => (
                  <li
                    key={i}
                    className='truncate rounded border bg-background/80 px-2 py-1'
                  >
                    {row}
                  </li>
                ))}
              </ul>
            </div>
          )}

          <div className='flex justify-end pt-1'>
            <Button
              type='button'
              size='sm'
              onClick={handleRouterImportExecute}
              disabled={isExecutingRouterImport}
              className='gap-1.5'
            >
              {isExecutingRouterImport ? (
                <>
                  <Loader2 className='h-4 w-4 animate-spin' />
                  <span>Menyimpan ke Database…</span>
                </>
              ) : (
                <>
                  <CheckCircle2 className='h-4 w-4' />
                  <span>
                    Konfirmasi & Simpan ({routerPreview.rowsTotal} Akun)
                  </span>
                </>
              )}
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
