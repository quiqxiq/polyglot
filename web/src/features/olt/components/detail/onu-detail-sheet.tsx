import { useState } from 'react'
import { Loader2 } from 'lucide-react'
import { useI18n } from '@/context/i18n-provider'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  useOnuEventsQuery,
  useOnuOpticalHistoryQuery,
  useOnuQuery,
} from '../../api/use-olt'
import {
  formatDateTime,
  formatDistance,
  formatRxPower,
  formatTemperature,
  positionLabel,
} from '../../lib/format'
import { OnuStateBadge, RxClassBadge } from '../olt-badges'

const HOUR_OPTIONS = [6, 24, 72, 168]

interface OnuDetailSheetProps {
  onuId: string | null
  onOpenChange: (open: boolean) => void
}

function InfoRow({
  label,
  children,
}: {
  label: string
  children: React.ReactNode
}) {
  return (
    <div className='flex items-start justify-between gap-4 border-b py-2 text-sm last:border-b-0'>
      <span className='shrink-0 text-muted-foreground'>{label}</span>
      <span className='text-right font-medium break-all'>{children}</span>
    </div>
  )
}

function LoadingBlock() {
  return (
    <div className='flex items-center justify-center py-10 text-muted-foreground'>
      <Loader2 className='h-5 w-5 animate-spin' />
    </div>
  )
}

export function OnuDetailSheet({ onuId, onOpenChange }: OnuDetailSheetProps) {
  const { t, language } = useI18n()
  const [hours, setHours] = useState(24)
  const { data: onu, isLoading } = useOnuQuery(onuId)
  const { data: samples = [], isLoading: isLoadingSamples } =
    useOnuOpticalHistoryQuery(onuId, hours)
  const { data: events = [], isLoading: isLoadingEvents } =
    useOnuEventsQuery(onuId)

  return (
    <Sheet open={onuId !== null} onOpenChange={onOpenChange}>
      <SheetContent className='w-full overflow-y-auto sm:max-w-xl'>
        <SheetHeader>
          <SheetTitle>
            {onu?.name ||
              (onu
                ? positionLabel(onu.ponIndex, onu.onuIndex)
                : t('olt.detail_sheet_title'))}
          </SheetTitle>
          <SheetDescription>{t('olt.detail_sheet_desc')}</SheetDescription>
        </SheetHeader>

        {isLoading || !onu ? (
          <LoadingBlock />
        ) : (
          <div className='px-4 pb-6'>
            <Tabs defaultValue='info'>
              <TabsList className='mb-3 w-full justify-start'>
                <TabsTrigger value='info'>{t('olt.info')}</TabsTrigger>
                <TabsTrigger value='optical'>{t('olt.rx_history')}</TabsTrigger>
                <TabsTrigger value='events'>{t('olt.events')}</TabsTrigger>
              </TabsList>

              <TabsContent value='info'>
                <div className='rounded-md border px-3'>
                  <InfoRow label={t('olt.state')}>
                    <OnuStateBadge state={onu.state} />
                  </InfoRow>
                  <InfoRow label={t('olt.admin_state')}>
                    <Badge variant='outline'>{onu.adminState || '-'}</Badge>
                  </InfoRow>
                  <InfoRow label={t('olt.pon')}>
                    <span className='font-mono'>
                      {positionLabel(onu.ponIndex, onu.onuIndex)}
                    </span>
                  </InfoRow>
                  <InfoRow label={t('olt.mac')}>
                    <span className='font-mono'>{onu.macAddress || '-'}</span>
                  </InfoRow>
                  <InfoRow label={t('olt.serial_number')}>
                    <span className='font-mono'>{onu.serialNumber || '-'}</span>
                  </InfoRow>
                  <InfoRow label={t('olt.model')}>{onu.model || '-'}</InfoRow>
                  <InfoRow label={t('olt.firmware')}>
                    {onu.firmware || '-'}
                  </InfoRow>
                  <InfoRow label={t('olt.rx_power')}>
                    <span className='font-mono'>
                      {formatRxPower(onu.rxPowerDbm)}
                    </span>{' '}
                    <RxClassBadge rxClass={onu.rxClass} />
                  </InfoRow>
                  <InfoRow label={t('olt.tx_power')}>
                    <span className='font-mono'>
                      {formatRxPower(onu.txPowerDbm)}
                    </span>
                  </InfoRow>
                  <InfoRow label={t('olt.distance')}>
                    {formatDistance(onu.distanceM)}
                  </InfoRow>
                  <InfoRow label={t('olt.temperature')}>
                    {formatTemperature(onu.temperatureC)}
                  </InfoRow>
                  <InfoRow label={t('olt.last_seen')}>
                    {formatDateTime(onu.lastSeenAt, language)}
                  </InfoRow>
                  <InfoRow label={t('olt.last_polled')}>
                    {formatDateTime(onu.lastPolledAt, language)}
                  </InfoRow>
                  <InfoRow label={t('olt.missed_polls')}>
                    {onu.missedPolls}
                  </InfoRow>
                  <InfoRow label={t('olt.customer')}>
                    {onu.customerId || '-'}
                  </InfoRow>
                  <InfoRow label={t('olt.pppoe_username')}>
                    {onu.pppoeUsername || '-'}
                  </InfoRow>
                </div>
              </TabsContent>

              <TabsContent value='optical'>
                <div className='mb-3 flex items-center justify-end gap-2'>
                  <span className='text-xs text-muted-foreground'>
                    {t('olt.rx_history_hours')}
                  </span>
                  <Select
                    value={String(hours)}
                    onValueChange={(value) => setHours(Number(value))}
                  >
                    <SelectTrigger className='h-8 w-28'>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {HOUR_OPTIONS.map((option) => (
                        <SelectItem key={option} value={String(option)}>
                          {option} h
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                {isLoadingSamples ? (
                  <LoadingBlock />
                ) : samples.length === 0 ? (
                  <p className='py-8 text-center text-sm text-muted-foreground'>
                    {t('olt.no_history')}
                  </p>
                ) : (
                  <div className='overflow-x-auto rounded-md border'>
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead>{t('common.time')}</TableHead>
                          <TableHead>Rx</TableHead>
                          <TableHead>Tx</TableHead>
                          <TableHead>{t('olt.temperature')}</TableHead>
                          <TableHead>{t('olt.distance')}</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {samples.map((sample, index) => (
                          <TableRow key={`${sample.sampledAt}-${index}`}>
                            <TableCell className='text-xs whitespace-nowrap'>
                              {formatDateTime(sample.sampledAt, language)}
                            </TableCell>
                            <TableCell className='font-mono text-xs'>
                              {formatRxPower(sample.rxPowerDbm)}
                            </TableCell>
                            <TableCell className='font-mono text-xs'>
                              {formatRxPower(sample.txPowerDbm)}
                            </TableCell>
                            <TableCell className='text-xs'>
                              {formatTemperature(sample.temperatureC)}
                            </TableCell>
                            <TableCell className='text-xs'>
                              {formatDistance(sample.distanceM)}
                            </TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </div>
                )}
              </TabsContent>

              <TabsContent value='events'>
                {isLoadingEvents ? (
                  <LoadingBlock />
                ) : events.length === 0 ? (
                  <p className='py-8 text-center text-sm text-muted-foreground'>
                    {t('olt.no_events')}
                  </p>
                ) : (
                  <ol className='space-y-3'>
                    {events.map((event) => (
                      <li
                        key={event.id}
                        className='rounded-md border p-3 text-sm'
                      >
                        <div className='flex items-center justify-between gap-2'>
                          <Badge variant='outline'>{event.eventType}</Badge>
                          <span className='text-xs text-muted-foreground'>
                            {formatDateTime(event.createdAt, language)}
                          </span>
                        </div>
                        {(event.oldState || event.newState) && (
                          <p className='mt-1 text-xs text-muted-foreground'>
                            {event.oldState || '-'} → {event.newState || '-'}
                          </p>
                        )}
                        {event.message && (
                          <p className='mt-1'>{event.message}</p>
                        )}
                      </li>
                    ))}
                  </ol>
                )}
              </TabsContent>
            </Tabs>
          </div>
        )}
      </SheetContent>
    </Sheet>
  )
}
