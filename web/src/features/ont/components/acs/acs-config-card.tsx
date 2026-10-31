import { Loader2, RadioTower, RefreshCw, Settings2 } from 'lucide-react'
import { toast } from 'sonner'
import { useI18n } from '@/context/i18n-provider'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import {
  useAcsConfigQuery,
  useEnsureAcsPresetMutation,
  useSyncAcsDevicesMutation,
} from '../../api/use-ont'

interface AcsConfigCardProps {
  canManage: boolean
}

function ConfigItem({
  label,
  value,
  mono,
}: {
  label: string
  value: React.ReactNode
  mono?: boolean
}) {
  return (
    <div className='space-y-0.5'>
      <p className='text-xs text-muted-foreground'>{label}</p>
      <p
        className={mono ? 'font-mono text-sm break-all' : 'text-sm font-medium'}
      >
        {value || '-'}
      </p>
    </div>
  )
}

export function AcsConfigCard({ canManage }: AcsConfigCardProps) {
  const { t } = useI18n()
  const { data: config, isLoading } = useAcsConfigQuery()
  const ensurePreset = useEnsureAcsPresetMutation()
  const syncDevices = useSyncAcsDevicesMutation()

  const handleEnsurePreset = async () => {
    try {
      const res = await ensurePreset.mutateAsync()
      toast.success(res.message || t('acs.ensure_success'))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('acs.ensure_failed'))
    }
  }

  const handleSync = async () => {
    try {
      const res = await syncDevices.mutateAsync()
      toast.success(t('ont.job_queued', { jobId: res.jobId || '-' }))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('acs.sync_failed'))
    }
  }

  return (
    <Card className='shadow-none'>
      <CardHeader>
        <div className='flex flex-wrap items-center justify-between gap-2'>
          <div className='flex items-center gap-2'>
            <RadioTower className='h-5 w-5 text-primary' />
            <CardTitle className='text-base'>{t('acs.config_title')}</CardTitle>
            {config && (
              <Badge
                variant='outline'
                className={
                  config.enabled
                    ? 'border-emerald-500/30 bg-emerald-500/15 text-emerald-700 dark:text-emerald-400'
                    : 'border-muted-foreground/30 bg-muted text-muted-foreground'
                }
              >
                {config.enabled ? t('common.enabled') : t('common.disabled')}
              </Badge>
            )}
          </div>
          {canManage && (
            <div className='flex flex-wrap items-center gap-2'>
              <Button
                size='sm'
                variant='outline'
                onClick={handleEnsurePreset}
                disabled={ensurePreset.isPending}
              >
                {ensurePreset.isPending ? (
                  <Loader2 className='h-4 w-4 animate-spin' />
                ) : (
                  <Settings2 className='h-4 w-4' />
                )}
                {t('acs.ensure_preset')}
              </Button>
              <Button
                size='sm'
                variant='outline'
                onClick={handleSync}
                disabled={syncDevices.isPending}
              >
                {syncDevices.isPending ? (
                  <Loader2 className='h-4 w-4 animate-spin' />
                ) : (
                  <RefreshCw className='h-4 w-4' />
                )}
                {t('acs.sync_devices')}
              </Button>
            </div>
          )}
        </div>
        <CardDescription>{t('acs.config_desc')}</CardDescription>
      </CardHeader>
      <CardContent className='space-y-4'>
        {isLoading ? (
          <div className='space-y-2'>
            <Skeleton className='h-5 w-full' />
            <Skeleton className='h-5 w-2/3' />
          </div>
        ) : !config ? (
          <p className='text-sm text-muted-foreground'>
            {t('acs.not_configured')}
          </p>
        ) : (
          <div className='grid gap-4 sm:grid-cols-2'>
            <ConfigItem
              label={t('acs.server_device_id')}
              value={config.serverDeviceId}
              mono
            />
            <ConfigItem label={t('acs.acs_url')} value={config.acsUrl} mono />
            <ConfigItem
              label={t('acs.periodic_inform_interval')}
              value={config.periodicInformInterval}
            />
            <ConfigItem
              label={t('acs.preset_name')}
              value={config.presetName}
              mono
            />
          </div>
        )}

        <p className='rounded-md border border-dashed p-2 text-xs text-muted-foreground'>
          {t('acs.edit_hint')}
        </p>
      </CardContent>
    </Card>
  )
}
