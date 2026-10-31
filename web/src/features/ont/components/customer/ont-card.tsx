import type { Onu } from '@/gen/v1/olt_pb'
import { Activity, Network, RotateCcw, Unlink, Wifi } from 'lucide-react'
import { useI18n } from '@/context/i18n-provider'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import {
  OnuStateBadge,
  RxClassBadge,
} from '@/features/olt/components/olt-badges'
import {
  formatDateTime,
  formatRxPower,
  positionLabel,
} from '@/features/olt/lib/format'

interface OntCardProps {
  onu: Onu
  canManage: boolean
  onRefreshTelemetry: (onu: Onu) => void
  onReboot: (onu: Onu) => void
  onSetWifi: (onu: Onu) => void
  onPushPppoe: (onu: Onu) => void
  onUnlink: (onu: Onu) => void
}

function OntMeta({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className='space-y-0.5'>
      <p className='text-[11px] text-muted-foreground'>{label}</p>
      <div className='text-xs font-medium break-all'>{value || '-'}</div>
    </div>
  )
}

export function OntCard({
  onu,
  canManage,
  onRefreshTelemetry,
  onReboot,
  onSetWifi,
  onPushPppoe,
  onUnlink,
}: OntCardProps) {
  const { t, language } = useI18n()

  return (
    <Card className='gap-4 py-4 shadow-none'>
      <CardContent className='space-y-3 px-4'>
        <div className='flex flex-wrap items-start justify-between gap-2'>
          <div>
            <p className='font-semibold'>
              {onu.name || positionLabel(onu.ponIndex, onu.onuIndex)}
            </p>
            <p className='font-mono text-[11px] text-muted-foreground'>
              {positionLabel(onu.ponIndex, onu.onuIndex)} ·{' '}
              {onu.macAddress || '-'}
            </p>
          </div>
          <OnuStateBadge state={onu.state} />
        </div>

        <div className='grid grid-cols-2 gap-3'>
          <OntMeta
            label={t('ont.rx')}
            value={
              <span className='flex flex-wrap items-center gap-1'>
                <span className='font-mono'>
                  {formatRxPower(onu.rxPowerDbm)}
                </span>
                <RxClassBadge rxClass={onu.rxClass} />
              </span>
            }
          />
          <OntMeta
            label={t('ont.acs_device')}
            value={
              <span className='font-mono'>
                {onu.acsDeviceId || (
                  <span className='text-muted-foreground'>
                    {t('common.none')}
                  </span>
                )}
              </span>
            }
          />
          <OntMeta
            label={t('ont.last_inform')}
            value={formatDateTime(onu.acsLastInformAt, language)}
          />
          <OntMeta
            label={t('ont.linked_pppoe')}
            value={onu.pppoeUsername || '-'}
          />
        </div>

        <div className='flex flex-wrap gap-1.5 border-t pt-3'>
          <Button
            size='sm'
            variant='outline'
            className='h-7 text-xs'
            disabled={!canManage}
            onClick={() => onRefreshTelemetry(onu)}
          >
            <Activity className='h-3.5 w-3.5' />
            {t('ont.refresh_telemetry')}
          </Button>
          <Button
            size='sm'
            variant='outline'
            className='h-7 text-xs'
            disabled={!canManage}
            onClick={() => onReboot(onu)}
          >
            <RotateCcw className='h-3.5 w-3.5' />
            {t('ont.reboot')}
          </Button>
          <Button
            size='sm'
            variant='outline'
            className='h-7 text-xs'
            disabled={!canManage}
            onClick={() => onSetWifi(onu)}
          >
            <Wifi className='h-3.5 w-3.5' />
            {t('ont.set_wifi')}
          </Button>
          <Button
            size='sm'
            variant='outline'
            className='h-7 text-xs'
            disabled={!canManage}
            onClick={() => onPushPppoe(onu)}
          >
            <Network className='h-3.5 w-3.5' />
            {t('ont.push_pppoe')}
          </Button>
          <Button
            size='sm'
            variant='outline'
            className='h-7 border-rose-500/30 text-xs text-rose-600 hover:text-rose-600'
            disabled={!canManage}
            onClick={() => onUnlink(onu)}
          >
            <Unlink className='h-3.5 w-3.5' />
            {t('ont.unlink')}
          </Button>
        </div>

        {onu.adminState && (
          <Badge variant='outline' className='text-[10px]'>
            {onu.adminState}
          </Badge>
        )}
      </CardContent>
    </Card>
  )
}
