import { useState } from 'react'
import type { AcsDevice } from '@/gen/v1/ont_pb'
import { Loader2, Pin, PinOff, RefreshCw, SearchIcon } from 'lucide-react'
import { toast } from 'sonner'
import { useI18n } from '@/context/i18n-provider'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import { Switch } from '@/components/ui/switch'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { formatDateTime } from '@/features/olt/lib/format'
import {
  useAcsDevicesQuery,
  usePinAcsDeviceMutation,
  useUnpinAcsDeviceMutation,
} from '../../api/use-ont'

interface AcsDevicesTableProps {
  canManage: boolean
}

function AcsDeviceRow({
  device,
  canManage,
}: {
  device: AcsDevice
  canManage: boolean
}) {
  const { t, language } = useI18n()
  const [ontId, setOntId] = useState(
    device.matchSource === 'MANUAL' ? device.matchedOntId : ''
  )
  const pinMutation = usePinAcsDeviceMutation()
  const unpinMutation = useUnpinAcsDeviceMutation()
  const isManual =
    device.matchSource === 'MANUAL' && Boolean(device.matchedOntId)

  const handlePin = async () => {
    try {
      const res = await pinMutation.mutateAsync({
        ontId: ontId.trim(),
        acsDeviceId: device.acsDeviceId,
      })
      toast.success(res.message || t('acs.pin_success'))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('acs.pin_failed'))
    }
  }

  const handleUnpin = async () => {
    try {
      const res = await unpinMutation.mutateAsync(device.matchedOntId)
      toast.success(res.message || t('acs.unpin_success'))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('acs.pin_failed'))
    }
  }

  return (
    <TableRow>
      <TableCell className='max-w-[14rem] truncate font-mono text-xs'>
        {device.acsDeviceId}
      </TableCell>
      <TableCell className='text-xs'>{device.productClass || '-'}</TableCell>
      <TableCell className='font-mono text-xs'>
        {device.serialNumber || '-'}
      </TableCell>
      <TableCell className='text-xs whitespace-nowrap text-muted-foreground'>
        {formatDateTime(device.lastInformAt, language)}
      </TableCell>
      <TableCell>
        <Badge
          variant='outline'
          className={
            device.matchSource
              ? 'border-indigo-500/30 bg-indigo-500/15 text-indigo-700 dark:text-indigo-400'
              : 'border-muted-foreground/30 bg-muted text-muted-foreground'
          }
        >
          {device.matchSource || '-'}
        </Badge>
      </TableCell>
      <TableCell className='font-mono text-xs'>
        {device.matchedOntId || '-'}
      </TableCell>
      <TableCell>
        <div className='flex items-center justify-end gap-1.5'>
          {isManual ? (
            <Button
              size='sm'
              variant='outline'
              className='h-7'
              onClick={handleUnpin}
              disabled={!canManage || unpinMutation.isPending}
            >
              {unpinMutation.isPending ? (
                <Loader2 className='h-3.5 w-3.5 animate-spin' />
              ) : (
                <PinOff className='h-3.5 w-3.5' />
              )}
              {t('acs.unpin')}
            </Button>
          ) : (
            <>
              <Input
                value={ontId}
                onChange={(e) => setOntId(e.target.value)}
                placeholder={t('acs.pin_ont_id')}
                className='h-7 w-36 font-mono text-xs'
                disabled={!canManage}
              />
              <Button
                size='sm'
                variant='outline'
                className='h-7'
                onClick={handlePin}
                disabled={!canManage || !ontId.trim() || pinMutation.isPending}
              >
                {pinMutation.isPending ? (
                  <Loader2 className='h-3.5 w-3.5 animate-spin' />
                ) : (
                  <Pin className='h-3.5 w-3.5' />
                )}
                {t('acs.pin')}
              </Button>
            </>
          )}
        </div>
      </TableCell>
    </TableRow>
  )
}

export function AcsDevicesTable({ canManage }: AcsDevicesTableProps) {
  const { t } = useI18n()
  const [onlyUnmatched, setOnlyUnmatched] = useState(false)
  const [search, setSearch] = useState('')
  const {
    data: devices = [],
    isLoading,
    isFetching,
    refetch,
  } = useAcsDevicesQuery(onlyUnmatched, search)

  return (
    <div className='space-y-3'>
      <div>
        <h3 className='text-lg font-medium'>{t('acs.devices_title')}</h3>
        <p className='text-sm text-muted-foreground'>{t('acs.devices_desc')}</p>
      </div>

      <div className='flex flex-wrap items-center gap-2'>
        <div className='relative w-full sm:max-w-xs'>
          <SearchIcon className='absolute start-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground' />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={t('acs.search_placeholder')}
            className='ps-8'
          />
        </div>
        <div className='flex items-center gap-2'>
          <Switch
            id='acs-only-unmatched'
            checked={onlyUnmatched}
            onCheckedChange={setOnlyUnmatched}
          />
          <Label htmlFor='acs-only-unmatched' className='text-sm'>
            {t('acs.only_unmatched')}
          </Label>
        </div>
        <Button
          variant='outline'
          size='sm'
          className='h-9'
          onClick={() => refetch()}
          disabled={isFetching}
        >
          <RefreshCw
            className={`mr-1.5 h-4 w-4 ${isFetching ? 'animate-spin' : ''}`}
          />
          {t('acs.refresh')}
        </Button>
      </div>

      {isLoading ? (
        <div className='space-y-2 rounded-md border p-4'>
          {Array.from({ length: 4 }).map((_, index) => (
            <Skeleton key={index} className='h-9 w-full' />
          ))}
        </div>
      ) : devices.length === 0 ? (
        <div className='rounded-md border py-10 text-center text-sm text-muted-foreground'>
          {t('acs.empty')}
        </div>
      ) : (
        <div className='overflow-x-auto rounded-md border'>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t('acs.device_id')}</TableHead>
                <TableHead>{t('acs.product_class')}</TableHead>
                <TableHead>{t('acs.serial')}</TableHead>
                <TableHead>{t('acs.last_inform')}</TableHead>
                <TableHead>{t('acs.match_source')}</TableHead>
                <TableHead>{t('acs.matched_ont_id')}</TableHead>
                <TableHead className='text-right'>{t('acs.pin')}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {devices.map((device) => (
                <AcsDeviceRow
                  key={device.id || device.acsDeviceId}
                  device={device}
                  canManage={canManage}
                />
              ))}
            </TableBody>
          </Table>
        </div>
      )}
    </div>
  )
}
