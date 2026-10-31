import { useState } from 'react'
import type { Onu } from '@/gen/v1/olt_pb'
import { Loader2, Wifi } from 'lucide-react'
import { toast } from 'sonner'
import { useI18n } from '@/context/i18n-provider'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { positionLabel } from '@/features/olt/lib/format'
import { useSetOntWifiMutation } from '../../api/use-ont'

interface OntWifiDialogProps {
  onu: Onu
  onClose: () => void
}

export function OntWifiDialog({ onu, onClose }: OntWifiDialogProps) {
  const { t } = useI18n()
  const [ssid, setSsid] = useState('')
  const [password, setPassword] = useState('')
  const [band, setBand] = useState('2.4')
  const wifiMutation = useSetOntWifiMutation()
  const position = positionLabel(onu.ponIndex, onu.onuIndex)

  const handleSubmit = async () => {
    if (!ssid.trim()) {
      toast.error(t('ont.wifi_ssid'))
      return
    }
    if (password.length > 0 && password.length < 8) {
      toast.error(t('ont.wifi_password_hint'))
      return
    }
    try {
      const res = await wifiMutation.mutateAsync({
        ontId: onu.id,
        ssid: ssid.trim(),
        password,
        band,
      })
      toast.success(t('ont.job_queued', { jobId: res.jobId || '-' }))
      onClose()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('ont.job_failed'))
    }
  }

  return (
    <Dialog
      open
      onOpenChange={(open) => {
        if (!open && !wifiMutation.isPending) onClose()
      }}
    >
      <DialogContent className='sm:max-w-md'>
        <DialogHeader className='text-start'>
          <DialogTitle className='flex items-center gap-2'>
            <Wifi className='h-5 w-5' />
            {t('ont.wifi_title')}
          </DialogTitle>
          <DialogDescription>
            {t('ont.wifi_desc', { position })}
          </DialogDescription>
        </DialogHeader>

        <div className='space-y-3 py-1'>
          <div className='space-y-1.5'>
            <Label htmlFor='ont-wifi-ssid'>{t('ont.wifi_ssid')}</Label>
            <Input
              id='ont-wifi-ssid'
              value={ssid}
              onChange={(e) => setSsid(e.target.value)}
              placeholder='Polyglot-Home'
            />
          </div>
          <div className='space-y-1.5'>
            <Label htmlFor='ont-wifi-password'>{t('ont.wifi_password')}</Label>
            <Input
              id='ont-wifi-password'
              type='password'
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder={t('ont.wifi_password_hint')}
            />
          </div>
          <div className='space-y-1.5'>
            <Label>{t('ont.wifi_band')}</Label>
            <Select value={band} onValueChange={setBand}>
              <SelectTrigger className='w-full'>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value='2.4'>{t('ont.wifi_band_24')}</SelectItem>
                <SelectItem value='5'>{t('ont.wifi_band_5')}</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        <DialogFooter>
          <Button
            variant='outline'
            onClick={onClose}
            disabled={wifiMutation.isPending}
          >
            {t('common.cancel')}
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={wifiMutation.isPending || !ssid.trim()}
          >
            {wifiMutation.isPending ? (
              <Loader2 className='h-4 w-4 animate-spin' />
            ) : (
              <Wifi className='h-4 w-4' />
            )}
            {t('common.save')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
