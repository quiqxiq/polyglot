import { useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { toast } from 'sonner'
import {
  Code,
  Copy,
  Check,
  Zap,
  Info,
  Send,
} from 'lucide-react'
import type { Device } from '@/gen/v1/device_pb'
import { ApplyRouterIntegrationScriptRequest } from '@/gen/v1/device_pb'
import {
  useRouterIntegrationScriptQuery,
  useApplyRouterIntegrationScriptMutation,
} from '../api/use-devices'

interface DeviceWebhookScriptsDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  device: Device | null
}

export function DeviceWebhookScriptsDialog({
  open,
  onOpenChange,
  device,
}: DeviceWebhookScriptsDialogProps) {
  const deviceId = device?.id || ''
  const { data: scripts, isLoading } = useRouterIntegrationScriptQuery(
    deviceId,
    'all',
    '',
    open && Boolean(deviceId)
  )

  const applyMutation = useApplyRouterIntegrationScriptMutation()

  const [copiedKey, setCopiedKey] = useState<string | null>(null)
  const [pppProfileTarget, setPppProfileTarget] = useState('default')
  const [hotspotProfileTarget, setHotspotProfileTarget] = useState('default')

  const copyToClipboard = (text: string, key: string) => {
    navigator.clipboard.writeText(text)
    setCopiedKey(key)
    toast.success('Script berhasil disalin ke clipboard')
    setTimeout(() => setCopiedKey(null), 2000)
  }

  const handleApplyPPP = async () => {
    if (!deviceId || !pppProfileTarget) return
    try {
      const res = await applyMutation.mutateAsync(
        new ApplyRouterIntegrationScriptRequest({
          deviceId,
          profileName: pppProfileTarget,
          serviceType: 'pppoe',
        })
      )
      toast.success(res.message || 'Script PPP berhasil diterapkan ke profile router!')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Gagal menerapkan script')
    }
  }

  const handleApplyHotspot = async () => {
    if (!deviceId || !hotspotProfileTarget) return
    try {
      const res = await applyMutation.mutateAsync(
        new ApplyRouterIntegrationScriptRequest({
          deviceId,
          profileName: hotspotProfileTarget,
          serviceType: 'hotspot',
        })
      )
      toast.success(res.message || 'Script Hotspot berhasil diterapkan ke profile router!')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Gagal menerapkan script')
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='max-w-2xl max-h-[90vh] overflow-y-auto'>
        <DialogHeader>
          <div className='flex items-center gap-2'>
            <div className='flex h-9 w-9 items-center justify-center rounded-lg bg-blue-500/10 text-blue-600 dark:text-blue-400'>
              <Code className='h-5 w-5' />
            </div>
            <div>
              <DialogTitle>Script Integrasi Webhook Router</DialogTitle>
              <DialogDescription>
                Script RouterOS untuk notifikasi real-time on-up/on-down PPPoE & on-login/on-logout Hotspot pada{' '}
                <span className='font-semibold text-foreground'>{device?.name || 'Router'}</span>
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <Alert className='bg-muted/50 border-border text-xs py-2.5'>
          <Info className='h-4 w-4 text-blue-500' />
          <AlertTitle className='font-semibold'>Fleksibilitas Integrasi</AlertTitle>
          <AlertDescription className='mt-1 text-muted-foreground'>
            Script ini opsional dan memberi Anda kontrol penuh. Anda dapat menyalinnya secara manual ke Winbox pada menu Profile, atau klik <strong>Terapkan Otomatis</strong> untuk memasangnya langsung.
          </AlertDescription>
        </Alert>

        <Tabs defaultValue='pppoe' className='w-full'>
          <TabsList className='grid grid-cols-2 w-full'>
            <TabsTrigger value='pppoe' className='gap-2 text-xs'>
              <Zap className='h-3.5 w-3.5' /> PPPoE (on-up & on-down)
            </TabsTrigger>
            <TabsTrigger value='hotspot' className='gap-2 text-xs'>
              <Zap className='h-3.5 w-3.5' /> Hotspot (on-login & on-logout)
            </TabsTrigger>
          </TabsList>

          {/* PPPoE Tab */}
          <TabsContent value='pppoe' className='space-y-4 pt-2'>
            <div className='flex items-center gap-3 rounded-lg border bg-muted/20 p-3'>
              <div className='flex-1 space-y-1'>
                <Label className='text-xs font-semibold'>Target PPP Profile di Router</Label>
                <Input
                  value={pppProfileTarget}
                  onChange={(e) => setPppProfileTarget(e.target.value)}
                  className='h-8 text-xs font-mono'
                  placeholder='default / profile-name'
                />
              </div>
              <div className='pt-5'>
                <Button
                  size='sm'
                  onClick={handleApplyPPP}
                  disabled={applyMutation.isPending || !pppProfileTarget}
                  className='h-8 gap-1 text-xs'
                >
                  <Send className='h-3.5 w-3.5' />
                  {applyMutation.isPending ? 'Menerapkan...' : 'Terapkan Otomatis'}
                </Button>
              </div>
            </div>

            <div className='space-y-2'>
              <div className='flex items-center justify-between'>
                <Label className='text-xs font-semibold text-muted-foreground'>
                  1. Script On-Up (/ppp profile on-up)
                </Label>
                <Button
                  variant='ghost'
                  size='sm'
                  onClick={() => copyToClipboard(scripts?.pppOnUpScript || '', 'ppp-up')}
                  className='h-7 text-xs gap-1'
                >
                  {copiedKey === 'ppp-up' ? <Check className='h-3.5 w-3.5 text-emerald-500' /> : <Copy className='h-3.5 w-3.5' />}
                  Salin Script
                </Button>
              </div>
              <pre className='rounded-lg bg-zinc-950 p-3 font-mono text-[11px] text-zinc-100 overflow-x-auto border border-zinc-800 leading-relaxed'>
                {isLoading ? 'Membuat script...' : scripts?.pppOnUpScript}
              </pre>
            </div>

            <div className='space-y-2'>
              <div className='flex items-center justify-between'>
                <Label className='text-xs font-semibold text-muted-foreground'>
                  2. Script On-Down (/ppp profile on-down)
                </Label>
                <Button
                  variant='ghost'
                  size='sm'
                  onClick={() => copyToClipboard(scripts?.pppOnDownScript || '', 'ppp-down')}
                  className='h-7 text-xs gap-1'
                >
                  {copiedKey === 'ppp-down' ? <Check className='h-3.5 w-3.5 text-emerald-500' /> : <Copy className='h-3.5 w-3.5' />}
                  Salin Script
                </Button>
              </div>
              <pre className='rounded-lg bg-zinc-950 p-3 font-mono text-[11px] text-zinc-100 overflow-x-auto border border-zinc-800 leading-relaxed'>
                {isLoading ? 'Membuat script...' : scripts?.pppOnDownScript}
              </pre>
            </div>
          </TabsContent>

          {/* Hotspot Tab */}
          <TabsContent value='hotspot' className='space-y-4 pt-2'>
            <div className='flex items-center gap-3 rounded-lg border bg-muted/20 p-3'>
              <div className='flex-1 space-y-1'>
                <Label className='text-xs font-semibold'>Target Hotspot User Profile di Router</Label>
                <Input
                  value={hotspotProfileTarget}
                  onChange={(e) => setHotspotProfileTarget(e.target.value)}
                  className='h-8 text-xs font-mono'
                  placeholder='default / profile-name'
                />
              </div>
              <div className='pt-5'>
                <Button
                  size='sm'
                  onClick={handleApplyHotspot}
                  disabled={applyMutation.isPending || !hotspotProfileTarget}
                  className='h-8 gap-1 text-xs'
                >
                  <Send className='h-3.5 w-3.5' />
                  {applyMutation.isPending ? 'Menerapkan...' : 'Terapkan Otomatis'}
                </Button>
              </div>
            </div>

            <div className='space-y-2'>
              <div className='flex items-center justify-between'>
                <Label className='text-xs font-semibold text-muted-foreground'>
                  1. Script On-Login (/ip hotspot user profile on-login)
                </Label>
                <Button
                  variant='ghost'
                  size='sm'
                  onClick={() => copyToClipboard(scripts?.hotspotOnLoginScript || '', 'hot-login')}
                  className='h-7 text-xs gap-1'
                >
                  {copiedKey === 'hot-login' ? <Check className='h-3.5 w-3.5 text-emerald-500' /> : <Copy className='h-3.5 w-3.5' />}
                  Salin Script
                </Button>
              </div>
              <pre className='rounded-lg bg-zinc-950 p-3 font-mono text-[11px] text-zinc-100 overflow-x-auto border border-zinc-800 leading-relaxed'>
                {isLoading ? 'Membuat script...' : scripts?.hotspotOnLoginScript}
              </pre>
            </div>

            <div className='space-y-2'>
              <div className='flex items-center justify-between'>
                <Label className='text-xs font-semibold text-muted-foreground'>
                  2. Script On-Logout (/ip hotspot user profile on-logout)
                </Label>
                <Button
                  variant='ghost'
                  size='sm'
                  onClick={() => copyToClipboard(scripts?.hotspotOnLogoutScript || '', 'hot-logout')}
                  className='h-7 text-xs gap-1'
                >
                  {copiedKey === 'hot-logout' ? <Check className='h-3.5 w-3.5 text-emerald-500' /> : <Copy className='h-3.5 w-3.5' />}
                  Salin Script
                </Button>
              </div>
              <pre className='rounded-lg bg-zinc-950 p-3 font-mono text-[11px] text-zinc-100 overflow-x-auto border border-zinc-800 leading-relaxed'>
                {isLoading ? 'Membuat script...' : scripts?.hotspotOnLogoutScript}
              </pre>
            </div>
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  )
}
