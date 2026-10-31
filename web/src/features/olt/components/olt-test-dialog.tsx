import type { OltDevice, TestOltConnectionResponse } from '@/gen/v1/olt_pb'
import { Loader2, RefreshCw, Zap } from 'lucide-react'
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
import { formatLatency } from '../lib/format'
import { HealthBadge } from './olt-badges'

interface OltTestDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  olt: OltDevice | null
  result: TestOltConnectionResponse | null
  isPending: boolean
  error: string | null
  onRetry: () => void
}

function HealthRow({
  label,
  value,
}: {
  label: string
  value: React.ReactNode
}) {
  return (
    <div className='flex items-center justify-between gap-4 text-sm'>
      <span className='text-muted-foreground'>{label}</span>
      <span className='text-right font-medium'>{value}</span>
    </div>
  )
}

export function OltTestDialog({
  open,
  onOpenChange,
  olt,
  result,
  isPending,
  error,
  onRetry,
}: OltTestDialogProps) {
  const { t } = useI18n()
  const health = result?.health

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader className='text-start'>
          <DialogTitle className='flex items-center gap-2'>
            <Zap className='h-5 w-5 text-amber-500' />
            {t('olt.test_title')}
          </DialogTitle>
          <DialogDescription>
            {t('olt.test_desc', {
              host: olt ? `${olt.host}:${olt.snmpPort}` : '-',
            })}
          </DialogDescription>
        </DialogHeader>

        {isPending ? (
          <div className='flex flex-col items-center justify-center gap-2 py-8'>
            <Loader2 className='h-8 w-8 animate-spin text-primary' />
            <p className='text-xs text-muted-foreground'>
              {t('olt.test_running')}
            </p>
          </div>
        ) : (
          <div className='space-y-3'>
            <div className='flex items-center justify-between rounded-lg border p-3 text-sm'>
              <span className='font-medium'>{t('olt.status')}</span>
              <HealthBadge health={health} />
            </div>

            {error && (
              <p className='rounded-md border border-rose-500/30 bg-rose-500/10 p-2 text-xs text-rose-700 dark:text-rose-400'>
                {error}
              </p>
            )}

            {health && (
              <div className='space-y-2 rounded-lg border p-3'>
                <HealthRow
                  label={t('olt.variant')}
                  value={health.variant || '-'}
                />
                <HealthRow
                  label={t('olt.firmware')}
                  value={health.firmware || '-'}
                />
                <HealthRow
                  label={t('olt.uptime')}
                  value={health.uptime || '-'}
                />
                <HealthRow
                  label={t('olt.latency')}
                  value={formatLatency(health.latencyMs)}
                />
                <HealthRow
                  label={t('olt.sys_descr')}
                  value={
                    <span className='line-clamp-3 font-mono text-xs'>
                      {health.sysDescr || '-'}
                    </span>
                  }
                />
              </div>
            )}
          </div>
        )}

        <DialogFooter>
          <Button variant='outline' onClick={() => onOpenChange(false)}>
            {t('common.close')}
          </Button>
          <Button onClick={onRetry} disabled={isPending}>
            {isPending ? (
              <Loader2 className='h-4 w-4 animate-spin' />
            ) : (
              <RefreshCw className='h-4 w-4' />
            )}
            {t('olt.test')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
