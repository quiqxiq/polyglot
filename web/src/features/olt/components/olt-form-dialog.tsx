import { useState } from 'react'
import type { OltDevice } from '@/gen/v1/olt_pb'
import { Loader2, Plus, Save } from 'lucide-react'
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
import { Switch } from '@/components/ui/switch'
import { useCreateOltMutation, useUpdateOltMutation } from '../api/use-olt'
import { OLT_TRANSPORT_OPTIONS, OLT_VARIANT_OPTIONS } from '../lib/constants'

interface OltFormDialogProps {
  olt?: OltDevice | null
  onClose: () => void
}

interface OltFormState {
  name: string
  host: string
  transport: string
  variant: string
  telnetPort: string
  sshPort: string
  snmpPort: string
  ponCount: string
  username: string
  password: string
  communityRo: string
  communityRw: string
  enabled: boolean
}

function initialState(olt?: OltDevice | null): OltFormState {
  return {
    name: olt?.name ?? '',
    host: olt?.host ?? '',
    transport: olt?.transport || 'telnet',
    variant:
      olt?.variant === 'HA7302' || olt?.variant === 'HA7304'
        ? olt.variant
        : 'AUTO',
    telnetPort: String(olt?.telnetPort || 23),
    sshPort: String(olt?.sshPort || 22),
    snmpPort: String(olt?.snmpPort || 161),
    ponCount: String(olt?.ponCount || 0),
    username: '',
    password: '',
    communityRo: '',
    communityRw: '',
    enabled: olt?.enabled ?? true,
  }
}

function parsePort(value: string): number {
  const parsed = Number.parseInt(value, 10)
  return Number.isFinite(parsed) && parsed > 0 ? parsed : 0
}

/**
 * Dialog create/update OLT. Komponen di-mount hanya saat terbuka agar state
 * form selalu mengikuti baris yang dipilih.
 */
export function OltFormDialog({ olt, onClose }: OltFormDialogProps) {
  const { t } = useI18n()
  const isUpdate = Boolean(olt)
  const [form, setForm] = useState<OltFormState>(() => initialState(olt))
  const createMutation = useCreateOltMutation()
  const updateMutation = useUpdateOltMutation()
  const isPending = createMutation.isPending || updateMutation.isPending

  const update = <K extends keyof OltFormState>(
    key: K,
    value: OltFormState[K]
  ) => setForm((prev) => ({ ...prev, [key]: value }))

  const handleSubmit = async () => {
    const host = form.host.trim()
    const communityRo = form.communityRo.trim()
    if (!host || !communityRo) {
      toast.error(t('olt.required_fields'))
      return
    }

    const payload = {
      name: form.name.trim(),
      host,
      transport: form.transport,
      variant: form.variant,
      telnetPort: parsePort(form.telnetPort),
      sshPort: parsePort(form.sshPort),
      snmpPort: parsePort(form.snmpPort),
      ponCount: parsePort(form.ponCount),
      username: form.username.trim(),
      password: form.password,
      communityRo,
      communityRw: form.communityRw.trim(),
      enabled: form.enabled,
    }

    try {
      if (olt) {
        await updateMutation.mutateAsync({ ...payload, id: olt.id })
        toast.success(t('olt.update_success'))
      } else {
        await createMutation.mutateAsync(payload)
        toast.success(t('olt.create_success'))
      }
      onClose()
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : t('alerts.action_failed')
      )
    }
  }

  return (
    <Dialog
      open
      onOpenChange={(open) => {
        if (!open && !isPending) onClose()
      }}
    >
      <DialogContent className='max-h-[90vh] overflow-y-auto sm:max-w-2xl'>
        <DialogHeader className='text-start'>
          <DialogTitle>
            {isUpdate ? t('olt.edit_title') : t('olt.create_title')}
          </DialogTitle>
          <DialogDescription>
            {isUpdate ? t('olt.edit_desc') : t('olt.create_desc')}
          </DialogDescription>
        </DialogHeader>

        <div className='grid gap-4 py-2 sm:grid-cols-2'>
          <div className='space-y-1.5'>
            <Label htmlFor='olt-name'>{t('olt.name')}</Label>
            <Input
              id='olt-name'
              value={form.name}
              onChange={(e) => update('name', e.target.value)}
              placeholder='OLT-POP-Kebonjeruk'
            />
          </div>
          <div className='space-y-1.5'>
            <Label htmlFor='olt-host'>{t('olt.host')} *</Label>
            <Input
              id='olt-host'
              value={form.host}
              onChange={(e) => update('host', e.target.value)}
              placeholder='10.10.10.2'
            />
          </div>

          <div className='space-y-1.5'>
            <Label>{t('olt.transport')}</Label>
            <Select
              value={form.transport}
              onValueChange={(value) => update('transport', value)}
            >
              <SelectTrigger className='w-full'>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {OLT_TRANSPORT_OPTIONS.map((option) => (
                  <SelectItem key={option.value} value={option.value}>
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className='space-y-1.5'>
            <Label>{t('olt.variant')}</Label>
            <Select
              value={form.variant}
              onValueChange={(value) => update('variant', value)}
            >
              <SelectTrigger className='w-full'>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {OLT_VARIANT_OPTIONS.map((option) => (
                  <SelectItem key={option} value={option}>
                    {option === 'AUTO' ? t('olt.variant_auto') : option}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className='space-y-1.5'>
            <Label htmlFor='olt-telnet-port'>{t('olt.telnet_port')}</Label>
            <Input
              id='olt-telnet-port'
              inputMode='numeric'
              value={form.telnetPort}
              onChange={(e) => update('telnetPort', e.target.value)}
            />
          </div>
          <div className='space-y-1.5'>
            <Label htmlFor='olt-ssh-port'>{t('olt.ssh_port')}</Label>
            <Input
              id='olt-ssh-port'
              inputMode='numeric'
              value={form.sshPort}
              onChange={(e) => update('sshPort', e.target.value)}
            />
          </div>
          <div className='space-y-1.5'>
            <Label htmlFor='olt-snmp-port'>{t('olt.snmp_port')}</Label>
            <Input
              id='olt-snmp-port'
              inputMode='numeric'
              value={form.snmpPort}
              onChange={(e) => update('snmpPort', e.target.value)}
            />
          </div>
          <div className='space-y-1.5'>
            <Label htmlFor='olt-pon-count'>{t('olt.pon_count')}</Label>
            <Input
              id='olt-pon-count'
              inputMode='numeric'
              value={form.ponCount}
              onChange={(e) => update('ponCount', e.target.value)}
            />
          </div>

          <div className='space-y-1.5'>
            <Label htmlFor='olt-username'>{t('olt.username')}</Label>
            <Input
              id='olt-username'
              autoComplete='off'
              value={form.username}
              onChange={(e) => update('username', e.target.value)}
            />
          </div>
          <div className='space-y-1.5'>
            <Label htmlFor='olt-password'>{t('olt.password')}</Label>
            <Input
              id='olt-password'
              type='password'
              autoComplete='new-password'
              value={form.password}
              onChange={(e) => update('password', e.target.value)}
            />
            {isUpdate && (
              <p className='text-xs text-muted-foreground'>
                {t('olt.password_keep_hint')}
              </p>
            )}
          </div>

          <div className='space-y-1.5'>
            <Label htmlFor='olt-community-ro'>{t('olt.community_ro')} *</Label>
            <Input
              id='olt-community-ro'
              value={form.communityRo}
              onChange={(e) => update('communityRo', e.target.value)}
              placeholder='public'
            />
          </div>
          <div className='space-y-1.5'>
            <Label htmlFor='olt-community-rw'>{t('olt.community_rw')}</Label>
            <Input
              id='olt-community-rw'
              value={form.communityRw}
              onChange={(e) => update('communityRw', e.target.value)}
              placeholder='private'
            />
          </div>
        </div>

        <p className='text-xs text-muted-foreground'>
          {t('olt.community_required_hint')}
        </p>

        <div className='flex items-center justify-between rounded-lg border p-3'>
          <div>
            <Label htmlFor='olt-enabled'>{t('olt.enabled')}</Label>
            <p className='text-xs text-muted-foreground'>{t('olt.subtitle')}</p>
          </div>
          <Switch
            id='olt-enabled'
            checked={form.enabled}
            onCheckedChange={(checked) => update('enabled', checked)}
          />
        </div>

        <DialogFooter>
          <Button
            type='button'
            variant='outline'
            onClick={onClose}
            disabled={isPending}
          >
            {t('common.cancel')}
          </Button>
          <Button type='button' onClick={handleSubmit} disabled={isPending}>
            {isPending ? (
              <Loader2 className='h-4 w-4 animate-spin' />
            ) : isUpdate ? (
              <Save className='h-4 w-4' />
            ) : (
              <Plus className='h-4 w-4' />
            )}
            {isUpdate ? t('common.save') : t('olt.add')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
