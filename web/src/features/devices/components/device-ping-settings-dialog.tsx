import { useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { toast } from 'sonner'
import type { Device, GetDevicePingConfigResponse } from '@/gen/v1/device_pb'
import {
  useDevicePingConfigQuery,
  useUpdateDevicePingConfigMutation,
} from '../api/use-devices'
import { getErrorMessage } from '../lib/formatters'
import { Activity, AlertTriangle, CheckCircle2, Database, Loader2 } from 'lucide-react'

interface DevicePingSettingsDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  device: Device | null
}

interface PingSettingsFormProps {
  device: Device
  configData?: GetDevicePingConfigResponse
  onOpenChange: (open: boolean) => void
}

function PingSettingsForm({
  device,
  configData,
  onOpenChange,
}: PingSettingsFormProps) {
  const updateMutation = useUpdateDevicePingConfigMutation()

  const [enabled, setEnabled] = useState<boolean>(() => {
    if (configData?.config) return configData.config.enabled
    return device.pingEnabled
  })

  const [target, setTarget] = useState<string>(() => {
    if (configData?.config?.target) return configData.config.target
    return device.pingTarget || '8.8.8.8'
  })

  const [retentionDays, setRetentionDays] = useState<number>(() => {
    if (configData?.config?.retentionDays) return configData.config.retentionDays
    return device.pingRetentionDays || 7
  })

  const timescaledbAvailable = configData?.timescaledbAvailable ?? true

  const handleSave = async () => {
    try {
      await updateMutation.mutateAsync({
        deviceId: device.id,
        enabled,
        target: target.trim() || '8.8.8.8',
        retentionDays,
      })
      toast.success(`Pengaturan ping untuk "${device.name}" berhasil disimpan`)
      onOpenChange(false)
    } catch (err: unknown) {
      toast.error(getErrorMessage(err, 'Gagal menyimpan konfigurasi ping'))
    }
  }

  return (
    <>
      <div className='space-y-4 py-2'>
        {/* TimescaleDB status notice */}
        <div className='flex items-center justify-between rounded-lg border p-3 bg-muted/30'>
          <div className='flex items-center gap-2'>
            <Database className='h-4 w-4 text-muted-foreground' />
            <span className='text-xs font-medium'>TimescaleDB Hypertable</span>
          </div>
          {timescaledbAvailable ? (
            <Badge
              variant='outline'
              className='bg-emerald-500/10 text-emerald-600 border-emerald-200 dark:border-emerald-800 text-[11px] gap-1'
            >
              <CheckCircle2 className='h-3 w-3' />
              Aktif
            </Badge>
          ) : (
            <Badge variant='destructive' className='text-[11px] gap-1'>
              <AlertTriangle className='h-3 w-3' />
              Belum Terpasang
            </Badge>
          )}
        </div>

        {!timescaledbAvailable && (
          <div className='rounded-md bg-amber-500/10 p-2.5 border border-amber-500/20 text-xs text-amber-700 dark:text-amber-400'>
            <span className='font-semibold'>Perhatian:</span> Ekstensi TimescaleDB belum aktif di PostgreSQL. Data metrik historis tidak akan disimpan hingga ekstensi diaktifkan.
          </div>
        )}

        {/* Toggle Enable/Disable */}
        <div className='flex items-center justify-between rounded-lg border p-3.5'>
          <div className='space-y-0.5'>
            <Label htmlFor='ping-enabled-switch' className='text-sm font-medium'>
              Aktifkan Pengumpulan Metrik Ping
            </Label>
            <p className='text-xs text-muted-foreground'>
              Streaming RouterOS akan otomatis merekam latensi & packet loss.
            </p>
          </div>
          <Switch
            id='ping-enabled-switch'
            checked={enabled}
            onCheckedChange={setEnabled}
          />
        </div>

        {/* Target Ping IP / Host */}
        <div className='space-y-1.5'>
          <Label htmlFor='ping-target-input' className='text-xs font-medium'>
            Target Ping (IP / Host / Gateway)
          </Label>
          <Input
            id='ping-target-input'
            placeholder='Contoh: 8.8.8.8 atau 192.168.1.1'
            value={target}
            onChange={(e) => setTarget(e.target.value)}
            className='font-mono text-xs'
          />
          <p className='text-[11px] text-muted-foreground'>
            Target default: <code className='text-primary font-mono'>8.8.8.8</code> (Google DNS) atau <code className='text-primary font-mono'>1.1.1.1</code> (Cloudflare).
          </p>
        </div>

        {/* Data Retention */}
        <div className='space-y-1.5'>
          <Label htmlFor='ping-retention-select' className='text-xs font-medium'>
            Retensi Penyimpanan Data (Hari)
          </Label>
          <Select
            value={String(retentionDays)}
            onValueChange={(val) => setRetentionDays(Number(val))}
          >
            <SelectTrigger id='ping-retention-select' className='text-xs'>
              <SelectValue placeholder='Pilih masa retensi' />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value='3'>3 Hari (Hemat Disk)</SelectItem>
              <SelectItem value='7'>7 Hari (Standar Rekomendasi)</SelectItem>
              <SelectItem value='14'>14 Hari (Dua Minggu)</SelectItem>
              <SelectItem value='30'>30 Hari (Satu Bulan)</SelectItem>
            </SelectContent>
          </Select>
          <p className='text-[11px] text-muted-foreground'>
            Data ping yang lebih tua dari batas hari ini akan dibersihkan otomatis.
          </p>
        </div>
      </div>

      <DialogFooter className='gap-2 sm:gap-0'>
        <Button
          variant='outline'
          size='sm'
          onClick={() => onOpenChange(false)}
          disabled={updateMutation.isPending}
        >
          Batal
        </Button>
        <Button
          size='sm'
          onClick={handleSave}
          disabled={updateMutation.isPending}
        >
          {updateMutation.isPending ? (
            <>
              <Loader2 className='mr-1.5 h-3.5 w-3.5 animate-spin' />
              Menyimpan...
            </>
          ) : (
            'Simpan Pengaturan'
          )}
        </Button>
      </DialogFooter>
    </>
  )
}

export function DevicePingSettingsDialog({
  open,
  onOpenChange,
  device,
}: DevicePingSettingsDialogProps) {
  const { data: configData, isLoading: isConfigLoading } =
    useDevicePingConfigQuery(device?.id || '', open && Boolean(device?.id))

  if (!device) return null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-[480px]'>
        <DialogHeader>
          <div className='flex items-center gap-2.5'>
            <div className='rounded-md bg-primary/10 p-2 text-primary'>
              <Activity className='h-5 w-5' />
            </div>
            <div>
              <DialogTitle className='text-base font-semibold'>
                Pengaturan Ping Metrics — {device.name}
              </DialogTitle>
              <DialogDescription className='text-xs text-muted-foreground'>
                Streaming telemetri ping realtime dan perekaman data historis
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        {isConfigLoading && !configData ? (
          <div className='flex h-36 items-center justify-center'>
            <Loader2 className='h-6 w-6 animate-spin text-muted-foreground' />
          </div>
        ) : (
          <PingSettingsForm
            key={`${device.id}_${configData?.config?.target || device.pingTarget || 'init'}`}
            device={device}
            configData={configData}
            onOpenChange={onOpenChange}
          />
        )}
      </DialogContent>
    </Dialog>
  )
}
