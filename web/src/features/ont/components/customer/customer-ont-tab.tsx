import { useState } from 'react'
import type { Onu } from '@/gen/v1/olt_pb'
import type { Subscription } from '@/gen/v1/subscription_pb'
import { Link2, Router } from 'lucide-react'
import { toast } from 'sonner'
import { useI18n } from '@/context/i18n-provider'
import { useCan } from '@/hooks/use-can'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { positionLabel } from '@/features/olt/lib/format'
import {
  useCustomerOnusQuery,
  useRebootOntMutation,
  useRefreshOntTelemetryMutation,
  useUnlinkOntMutation,
} from '../../api/use-ont'
import { OntCard } from './ont-card'
import { OntLinkDialog } from './ont-link-dialog'
import { OntPppoeDialog } from './ont-pppoe-dialog'
import { OntWifiDialog } from './ont-wifi-dialog'

interface CustomerOntTabProps {
  customerId: string
  subscriptions: Subscription[]
}

export function CustomerOntTab({
  customerId,
  subscriptions,
}: CustomerOntTabProps) {
  const { t } = useI18n()
  const canManage = useCan('ont:manage')
  const canReadOlt = useCan('olt:read')
  const {
    data: onus = [],
    isLoading,
    isError,
  } = useCustomerOnusQuery(customerId, canReadOlt)

  const telemetryMutation = useRefreshOntTelemetryMutation()
  const rebootMutation = useRebootOntMutation()
  const unlinkMutation = useUnlinkOntMutation()

  const [linkOpen, setLinkOpen] = useState(false)
  const [wifiOnu, setWifiOnu] = useState<Onu | null>(null)
  const [pppoeOnu, setPppoeOnu] = useState<Onu | null>(null)
  const [rebootOnu, setRebootOnu] = useState<Onu | null>(null)
  const [unlinkOnu, setUnlinkOnu] = useState<Onu | null>(null)

  const runJob = async (task: () => Promise<{ jobId: string }>) => {
    try {
      const res = await task()
      toast.success(t('ont.job_queued', { jobId: res.jobId || '-' }))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('ont.job_failed'))
    }
  }

  const handleReboot = async () => {
    if (!rebootOnu) return
    const onu = rebootOnu
    setRebootOnu(null)
    await runJob(() => rebootMutation.mutateAsync(onu.id))
  }

  const handleUnlink = async () => {
    if (!unlinkOnu) return
    try {
      const res = await unlinkMutation.mutateAsync(unlinkOnu.id)
      toast.success(res.message || t('ont.unlink_success'))
      setUnlinkOnu(null)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('ont.unlink_failed'))
    }
  }

  return (
    <div className='space-y-3'>
      <div className='flex flex-wrap items-end justify-between gap-2'>
        <div>
          <h3 className='text-lg font-medium'>{t('ont.title')}</h3>
          <p className='text-sm text-muted-foreground'>{t('ont.subtitle')}</p>
        </div>
        {canManage && canReadOlt && (
          <Button size='sm' className='h-8' onClick={() => setLinkOpen(true)}>
            <Link2 className='h-4 w-4' />
            {t('ont.link')}
          </Button>
        )}
      </div>

      {!canReadOlt ? (
        <Card className='border-dashed shadow-none'>
          <CardContent className='py-8 text-center text-sm text-muted-foreground'>
            {t('ont.no_access')}
          </CardContent>
        </Card>
      ) : isLoading ? (
        <div className='grid gap-3 lg:grid-cols-2'>
          {Array.from({ length: 2 }).map((_, index) => (
            <Skeleton key={index} className='h-44 w-full rounded-xl' />
          ))}
        </div>
      ) : isError ? (
        <Card className='border-dashed shadow-none'>
          <CardContent className='py-8 text-center text-sm text-muted-foreground'>
            {t('ont.load_failed')}
          </CardContent>
        </Card>
      ) : onus.length === 0 ? (
        <Card className='border-dashed shadow-none'>
          <CardContent className='flex flex-col items-center py-10 text-center text-sm text-muted-foreground'>
            <Router className='mb-2 size-8 opacity-60' />
            {t('ont.empty')}
          </CardContent>
        </Card>
      ) : (
        <div className='grid gap-3 lg:grid-cols-2'>
          {onus.map((onu) => (
            <OntCard
              key={onu.id}
              onu={onu}
              canManage={canManage}
              onRefreshTelemetry={(item) =>
                void runJob(() => telemetryMutation.mutateAsync(item.id))
              }
              onReboot={setRebootOnu}
              onSetWifi={setWifiOnu}
              onPushPppoe={setPppoeOnu}
              onUnlink={setUnlinkOnu}
            />
          ))}
        </div>
      )}

      {linkOpen && (
        <OntLinkDialog
          customerId={customerId}
          subscriptions={subscriptions}
          onClose={() => setLinkOpen(false)}
        />
      )}

      {wifiOnu && (
        <OntWifiDialog onu={wifiOnu} onClose={() => setWifiOnu(null)} />
      )}

      {pppoeOnu && (
        <OntPppoeDialog onu={pppoeOnu} onClose={() => setPppoeOnu(null)} />
      )}

      <ConfirmDialog
        open={rebootOnu !== null}
        onOpenChange={(open) => {
          if (!open) setRebootOnu(null)
        }}
        title={t('ont.reboot_title')}
        desc={
          rebootOnu
            ? t('ont.reboot_confirm', {
                position: positionLabel(rebootOnu.ponIndex, rebootOnu.onuIndex),
              })
            : ''
        }
        confirmText={t('ont.reboot')}
        isLoading={rebootMutation.isPending}
        handleConfirm={handleReboot}
      />

      <ConfirmDialog
        open={unlinkOnu !== null}
        onOpenChange={(open) => {
          if (!open) setUnlinkOnu(null)
        }}
        title={t('ont.unlink_title')}
        desc={
          unlinkOnu
            ? t('ont.unlink_confirm', {
                position: positionLabel(unlinkOnu.ponIndex, unlinkOnu.onuIndex),
              })
            : ''
        }
        confirmText={t('ont.unlink')}
        destructive
        isLoading={unlinkMutation.isPending}
        handleConfirm={handleUnlink}
      />
    </div>
  )
}
