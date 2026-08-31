import { useState, useEffect } from 'react'
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
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { Separator } from '@/components/ui/separator'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Textarea } from '@/components/ui/textarea'
import { toast } from 'sonner'
import {
  ShieldAlert,
  ShieldCheck,
  CheckCircle2,
  XCircle,
  RefreshCw,
  Server,
  Sparkles,
  ExternalLink,
  Globe,
} from 'lucide-react'
import type { Device } from '@/gen/v1/device_pb'
import { IsolationConfig, CreateIsolationProfileRequest } from '@/gen/v1/device_pb'
import {
  useIsolationStatusQuery,
  useCreateIsolationProfileMutation,
} from '../api/use-devices'

interface DeviceIsolationDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  device: Device | null
}

export function DeviceIsolationDialog({
  open,
  onOpenChange,
  device,
}: DeviceIsolationDialogProps) {
  const deviceId = device?.id || ''
  const {
    data: statusRes,
    isLoading: isStatusLoading,
    refetch,
  } = useIsolationStatusQuery(deviceId, open && Boolean(deviceId))

  const createMutation = useCreateIsolationProfileMutation()

  const [pppoeProfile, setPppoeProfile] = useState('ISOLIR')
  const [hotspotProfile, setHotspotProfile] = useState('ISOLIR')
  const [addressList, setAddressList] = useState('ISOLIR_USERS')
  const [rateLimit, setRateLimit] = useState('0/0')
  const [pppoeRedirectUrl, setPppoeRedirectUrl] = useState('/portal/isolate/pppoe')
  const [hotspotRedirectUrl, setHotspotRedirectUrl] = useState('/portal/isolate/hotspot')
  const [walledGardenText, setWalledGardenText] = useState('tripay.co.id, api.tripay.co.id, midtrans.com, app.midtrans.com, whatsapp.com')
  const [redirectIp, setRedirectIp] = useState(device?.host || '192.168.88.1')
  const [redirectPort, setRedirectPort] = useState(80)
  const [natRedirectEnabled, setNatRedirectEnabled] = useState(true)

  useEffect(() => {
    if (statusRes?.status?.config) {
      const c = statusRes.status.config
      if (c.pppoeProfileName) setPppoeProfile(c.pppoeProfileName)
      if (c.hotspotProfileName) setHotspotProfile(c.hotspotProfileName)
      if (c.addressListName) setAddressList(c.addressListName)
      if (c.rateLimit) setRateLimit(c.rateLimit)
      if (c.pppoeRedirectUrl) setPppoeRedirectUrl(c.pppoeRedirectUrl)
      if (c.hotspotRedirectUrl) setHotspotRedirectUrl(c.hotspotRedirectUrl)
      if (c.walledGardenDomains && c.walledGardenDomains.length > 0) {
        setWalledGardenText(c.walledGardenDomains.join(', '))
      }
      if (c.redirectIp) setRedirectIp(c.redirectIp)
      if (c.redirectPort) setRedirectPort(c.redirectPort)
      setNatRedirectEnabled(c.natRedirectEnabled)
    }
  }, [statusRes])

  const status = statusRes?.status

  const handleApply = async () => {
    if (!deviceId) return
    try {
      const domains = walledGardenText
        .split(',')
        .map((d) => d.trim())
        .filter(Boolean)

      const cfg = new IsolationConfig({
        pppoeProfileName: pppoeProfile,
        hotspotProfileName: hotspotProfile,
        addressListName: addressList,
        rateLimit: rateLimit,
        pppoeRedirectUrl: pppoeRedirectUrl,
        hotspotRedirectUrl: hotspotRedirectUrl,
        walledGardenDomains: domains,
        redirectIp: redirectIp,
        redirectPort: redirectPort,
        natRedirectEnabled: natRedirectEnabled,
      })

      const res = await createMutation.mutateAsync(
        new CreateIsolationProfileRequest({
          deviceId,
          config: cfg,
        })
      )
      toast.success(res.message || 'Profil isolir & firewall berhasil dikonfigurasi di router!')
      refetch()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Gagal memasang profil isolir')
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='max-w-xl max-h-[90vh] overflow-y-auto'>
        <DialogHeader>
          <div className='flex items-center gap-2'>
            <div className='flex h-9 w-9 items-center justify-center rounded-lg bg-amber-500/10 text-amber-600 dark:text-amber-400'>
              <ShieldAlert className='h-5 w-5' />
            </div>
            <div>
              <DialogTitle>Manajemen Profil Isolir Router</DialogTitle>
              <DialogDescription>
                Konfigurasi profil isolir PPPoE/Hotspot, custom redirect URL portal, dan firewall untuk{' '}
                <span className='font-semibold text-foreground'>{device?.name || 'Router'}</span>
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        {/* Live Router Status */}
        <div className='rounded-lg border bg-muted/40 p-4 space-y-3'>
          <div className='flex items-center justify-between'>
            <div className='flex items-center gap-2 font-medium text-sm'>
              <Server className='h-4 w-4 text-muted-foreground' />
              Status Infrastruktur di Router
            </div>
            <Button
              variant='ghost'
              size='sm'
              onClick={() => refetch()}
              disabled={isStatusLoading}
              className='h-7 px-2 text-xs gap-1'
            >
              <RefreshCw className={`h-3.5 w-3.5 ${isStatusLoading ? 'animate-spin' : ''}`} />
              Refresh
            </Button>
          </div>

          <div className='grid grid-cols-2 gap-3 text-xs'>
            <div className='flex items-center justify-between rounded-md border bg-background p-2.5'>
              <span className='text-muted-foreground'>PPP Profile ({pppoeProfile})</span>
              {status?.pppoeProfileExists ? (
                <Badge variant='outline' className='bg-emerald-500/15 text-emerald-700 border-emerald-500/30 gap-1'>
                  <CheckCircle2 className='h-3 w-3' /> Ada
                </Badge>
              ) : (
                <Badge variant='outline' className='bg-zinc-500/15 text-zinc-500 gap-1'>
                  <XCircle className='h-3 w-3' /> Belum Ada
                </Badge>
              )}
            </div>

            <div className='flex items-center justify-between rounded-md border bg-background p-2.5'>
              <span className='text-muted-foreground'>Hotspot Profile ({hotspotProfile})</span>
              {status?.hotspotProfileExists ? (
                <Badge variant='outline' className='bg-emerald-500/15 text-emerald-700 border-emerald-500/30 gap-1'>
                  <CheckCircle2 className='h-3 w-3' /> Ada
                </Badge>
              ) : (
                <Badge variant='outline' className='bg-zinc-500/15 text-zinc-500 gap-1'>
                  <XCircle className='h-3 w-3' /> Belum Ada
                </Badge>
              )}
            </div>

            <div className='flex items-center justify-between rounded-md border bg-background p-2.5'>
              <span className='text-muted-foreground'>Firewall NAT Redirect</span>
              {status?.natRedirectExists ? (
                <Badge variant='outline' className='bg-emerald-500/15 text-emerald-700 border-emerald-500/30 gap-1'>
                  <CheckCircle2 className='h-3 w-3' /> Aktif
                </Badge>
              ) : (
                <Badge variant='outline' className='bg-zinc-500/15 text-zinc-500 gap-1'>
                  <XCircle className='h-3 w-3' /> Belum Ada
                </Badge>
              )}
            </div>

            <div className='flex items-center justify-between rounded-md border bg-background p-2.5'>
              <span className='text-muted-foreground'>Pelanggan Terisolir</span>
              <Badge variant='secondary' className='font-mono'>
                {status?.isolatedUsersCount || 0} user
              </Badge>
            </div>
          </div>
        </div>

        <Alert className='bg-blue-500/10 border-blue-500/20 text-xs py-2.5'>
          <Sparkles className='h-4 w-4 text-blue-600 dark:text-blue-400' />
          <AlertTitle className='font-semibold text-blue-900 dark:text-blue-300'>
            Provisi Otomatis & Portal Pembayaran
          </AlertTitle>
          <AlertDescription className='text-blue-700 dark:text-blue-400 mt-1'>
            Klik tombol di bawah untuk membuat PPP profile ISOLIR, Hotspot user profile ISOLIR, dan firewall redirect ke halaman portal isolir Polyglot atau custom domain Anda.
          </AlertDescription>
        </Alert>

        <Separator />

        {/* Configuration Inputs */}
        <div className='space-y-4 text-xs'>
          {/* Section: URL Landing Portal Terpisah */}
          <div className='space-y-3 rounded-lg border bg-muted/20 p-3'>
            <div className='flex items-center gap-1.5 font-semibold text-foreground text-xs'>
              <Globe className='h-3.5 w-3.5 text-blue-500' />
              URL Landing Portal Isolir & Pembayaran
            </div>

            <div className='space-y-1.5'>
              <div className='flex items-center justify-between'>
                <Label className='text-xs'>1. URL Halaman Isolir PPPoE</Label>
                <a
                  href={pppoeRedirectUrl}
                  target='_blank'
                  rel='noreferrer'
                  className='text-[11px] text-blue-600 dark:text-blue-400 hover:underline flex items-center gap-1'
                >
                  Buka Halaman <ExternalLink className='h-3 w-3' />
                </a>
              </div>
              <Input
                value={pppoeRedirectUrl}
                onChange={(e) => setPppoeRedirectUrl(e.target.value)}
                className='h-8 font-mono text-xs'
                placeholder='/portal/isolate/pppoe atau https://tagihan.isp.id/isolated'
              />
            </div>

            <div className='space-y-1.5'>
              <div className='flex items-center justify-between'>
                <Label className='text-xs'>2. URL Halaman Isolir Hotspot</Label>
                <a
                  href={hotspotRedirectUrl}
                  target='_blank'
                  rel='noreferrer'
                  className='text-[11px] text-blue-600 dark:text-blue-400 hover:underline flex items-center gap-1'
                >
                  Buka Halaman <ExternalLink className='h-3 w-3' />
                </a>
              </div>
              <Input
                value={hotspotRedirectUrl}
                onChange={(e) => setHotspotRedirectUrl(e.target.value)}
                className='h-8 font-mono text-xs'
                placeholder='/portal/isolate/hotspot atau https://hotspot.isp.id/expired'
              />
            </div>

            <div className='space-y-1.5 pt-1'>
              <Label className='text-xs'>Domain Whitelist Payment Gateway (Walled Garden)</Label>
              <Textarea
                value={walledGardenText}
                onChange={(e) => setWalledGardenText(e.target.value)}
                rows={2}
                className='font-mono text-xs resize-none'
                placeholder='tripay.co.id, api.tripay.co.id, midtrans.com, whatsapp.com'
              />
              <p className='text-[11px] text-muted-foreground'>
                Domain yang diizinkan diakses pelanggan terisolir agar dapat memuat QRIS & Virtual Account. Pisahkan dengan koma.
              </p>
            </div>
          </div>

          <div className='grid grid-cols-2 gap-4'>
            <div className='space-y-1.5'>
              <Label className='text-xs'>Nama Profil PPPoE</Label>
              <Input
                value={pppoeProfile}
                onChange={(e) => setPppoeProfile(e.target.value)}
                className='h-8 font-mono text-xs'
                placeholder='ISOLIR'
              />
            </div>
            <div className='space-y-1.5'>
              <Label className='text-xs'>Nama Profil Hotspot</Label>
              <Input
                value={hotspotProfile}
                onChange={(e) => setHotspotProfile(e.target.value)}
                className='h-8 font-mono text-xs'
                placeholder='ISOLIR'
              />
            </div>
          </div>

          <div className='grid grid-cols-2 gap-4'>
            <div className='space-y-1.5'>
              <Label className='text-xs'>Address List Isolir</Label>
              <Input
                value={addressList}
                onChange={(e) => setAddressList(e.target.value)}
                className='h-8 font-mono text-xs'
                placeholder='ISOLIR_USERS'
              />
            </div>
            <div className='space-y-1.5'>
              <Label className='text-xs'>Rate Limit Isolir</Label>
              <Input
                value={rateLimit}
                onChange={(e) => setRateLimit(e.target.value)}
                className='h-8 font-mono text-xs'
                placeholder='0/0'
              />
            </div>
          </div>

          <div className='grid grid-cols-3 gap-4 items-end'>
            <div className='col-span-2 space-y-1.5'>
              <Label className='text-xs'>IP Target L3 Redirect</Label>
              <Input
                value={redirectIp}
                onChange={(e) => setRedirectIp(e.target.value)}
                className='h-8 font-mono text-xs'
                placeholder='192.168.88.1'
              />
            </div>
            <div className='space-y-1.5'>
              <Label className='text-xs'>Port Target</Label>
              <Input
                type='number'
                value={redirectPort}
                onChange={(e) => setRedirectPort(parseInt(e.target.value) || 80)}
                className='h-8 font-mono text-xs'
                placeholder='80'
              />
            </div>
          </div>

          <div className='flex items-center justify-between rounded-lg border p-3'>
            <div className='space-y-0.5'>
              <Label className='text-xs font-semibold'>Firewall NAT Redirect (dst-nat)</Label>
              <p className='text-[11px] text-muted-foreground'>
                Redirect traffic HTTP pelanggan terisolir ke portal pembayaran.
              </p>
            </div>
            <Switch
              checked={natRedirectEnabled}
              onCheckedChange={setNatRedirectEnabled}
            />
          </div>
        </div>

        <DialogFooter className='gap-2 sm:gap-0'>
          <Button variant='outline' onClick={() => onOpenChange(false)}>
            Tutup
          </Button>
          <Button
            onClick={handleApply}
            disabled={createMutation.isPending}
            className='gap-1.5 bg-amber-600 hover:bg-amber-700 text-white'
          >
            <ShieldCheck className='h-4 w-4' />
            {createMutation.isPending ? 'Menerapkan...' : 'Pasang / Terapkan di Router'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
