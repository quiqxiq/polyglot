import { useState } from 'react'
import type { Subscription } from '@/gen/v1/subscription_pb'
import { Link2, Loader2 } from 'lucide-react'
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
import { useOltsQuery } from '@/features/olt/api/use-olt'
import { positionLabel } from '@/features/olt/lib/format'
import { useLinkOntMutation, useUnlinkedOnusQuery } from '../../api/use-ont'

const NONE_VALUE = 'NONE'

interface OntLinkDialogProps {
  customerId: string
  subscriptions: Subscription[]
  onClose: () => void
}

export function OntLinkDialog({
  customerId,
  subscriptions,
  onClose,
}: OntLinkDialogProps) {
  const { t } = useI18n()
  const [oltId, setOltId] = useState('')
  const [onuId, setOnuId] = useState('')
  const [subscriptionId, setSubscriptionId] = useState('')
  const [pppoeUsername, setPppoeUsername] = useState('')

  const { data: olts = [], isLoading: isLoadingOlts } = useOltsQuery()
  const { data: unlinkedOnus = [], isLoading: isLoadingOnus } =
    useUnlinkedOnusQuery(oltId)
  const linkMutation = useLinkOntMutation()

  const handleSubscriptionChange = (value: string) => {
    const next = value === NONE_VALUE ? '' : value
    setSubscriptionId(next)
    const subscription = subscriptions.find((item) => item.id === next)
    if (subscription?.remoteUsername) {
      setPppoeUsername(subscription.remoteUsername)
    }
  }

  const handleSubmit = async () => {
    if (!onuId) {
      toast.error(t('ont.link_failed'))
      return
    }
    try {
      const res = await linkMutation.mutateAsync({
        ontId: onuId,
        customerId,
        subscriptionId,
        pppoeUsername: pppoeUsername.trim(),
      })
      toast.success(res.message || t('ont.link_success'))
      onClose()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('ont.link_failed'))
    }
  }

  return (
    <Dialog
      open
      onOpenChange={(open) => {
        if (!open && !linkMutation.isPending) onClose()
      }}
    >
      <DialogContent className='sm:max-w-lg'>
        <DialogHeader className='text-start'>
          <DialogTitle className='flex items-center gap-2'>
            <Link2 className='h-5 w-5' />
            {t('ont.link_title')}
          </DialogTitle>
          <DialogDescription>{t('ont.link_desc')}</DialogDescription>
        </DialogHeader>

        <div className='space-y-3 py-1'>
          <div className='space-y-1.5'>
            <Label>{t('ont.select_olt')}</Label>
            <Select
              value={oltId || undefined}
              onValueChange={(value) => {
                setOltId(value)
                setOnuId('')
              }}
              disabled={isLoadingOlts}
            >
              <SelectTrigger className='w-full'>
                <SelectValue placeholder={t('ont.select_olt')} />
              </SelectTrigger>
              <SelectContent>
                {olts.map((olt) => (
                  <SelectItem key={olt.id} value={olt.id}>
                    {olt.name || olt.host} ({olt.host})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className='space-y-1.5'>
            <Label>{t('ont.select_ont')}</Label>
            <Select
              value={onuId || undefined}
              onValueChange={setOnuId}
              disabled={!oltId || isLoadingOnus}
            >
              <SelectTrigger className='w-full'>
                <SelectValue
                  placeholder={
                    isLoadingOnus ? t('common.loading') : t('ont.select_ont')
                  }
                />
              </SelectTrigger>
              <SelectContent>
                {unlinkedOnus.map((onu) => (
                  <SelectItem key={onu.id} value={onu.id}>
                    {positionLabel(onu.ponIndex, onu.onuIndex)}
                    {onu.name ? ` · ${onu.name}` : ''}
                    {onu.macAddress ? ` · ${onu.macAddress}` : ''}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {oltId && !isLoadingOnus && unlinkedOnus.length === 0 && (
              <p className='text-xs text-muted-foreground'>
                {t('ont.unlinked_empty')}
              </p>
            )}
          </div>

          <div className='space-y-1.5'>
            <Label>{t('ont.select_subscription')}</Label>
            <Select
              value={subscriptionId || NONE_VALUE}
              onValueChange={handleSubscriptionChange}
            >
              <SelectTrigger className='w-full'>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={NONE_VALUE}>
                  {t('ont.no_subscription')}
                </SelectItem>
                {subscriptions.map((subscription) => (
                  <SelectItem key={subscription.id} value={subscription.id}>
                    {subscription.planName || subscription.id} (
                    {subscription.status})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className='space-y-1.5'>
            <Label htmlFor='ont-link-pppoe'>{t('ont.pppoe_username')}</Label>
            <Input
              id='ont-link-pppoe'
              value={pppoeUsername}
              onChange={(e) => setPppoeUsername(e.target.value)}
              placeholder='cust01'
            />
          </div>
        </div>

        <DialogFooter>
          <Button
            variant='outline'
            onClick={onClose}
            disabled={linkMutation.isPending}
          >
            {t('common.cancel')}
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={linkMutation.isPending || !onuId}
          >
            {linkMutation.isPending ? (
              <Loader2 className='h-4 w-4 animate-spin' />
            ) : (
              <Link2 className='h-4 w-4' />
            )}
            {t('ont.link')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
