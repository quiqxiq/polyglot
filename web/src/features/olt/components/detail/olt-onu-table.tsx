import type { Onu } from '@/gen/v1/olt_pb'
import {
  MoreHorizontal,
  Pencil,
  Power,
  PowerOff,
  RotateCcw,
  Trash2,
} from 'lucide-react'
import { useI18n } from '@/context/i18n-provider'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { rxClassTextClass } from '../../lib/constants'
import {
  formatDateTime,
  formatDistance,
  formatRxPower,
  positionLabel,
} from '../../lib/format'
import { OnuStateBadge, RxClassBadge } from '../olt-badges'

interface OltOnuTableProps {
  onus: Onu[]
  isLoading: boolean
  onSelect: (onu: Onu) => void
  onRename: (onu: Onu) => void
  onReboot: (onu: Onu) => void
  onToggleAdmin: (onu: Onu) => void
  onDelete: (onu: Onu) => void
}

export function OltOnuTable({
  onus,
  isLoading,
  onSelect,
  onRename,
  onReboot,
  onToggleAdmin,
  onDelete,
}: OltOnuTableProps) {
  const { t, language } = useI18n()

  if (isLoading) {
    return (
      <div className='space-y-2 rounded-md border p-4'>
        {Array.from({ length: 6 }).map((_, index) => (
          <Skeleton key={index} className='h-10 w-full' />
        ))}
      </div>
    )
  }

  if (onus.length === 0) {
    return (
      <div className='rounded-md border py-12 text-center text-sm text-muted-foreground'>
        {t('olt.empty_onus')}
      </div>
    )
  }

  return (
    <div className='overflow-x-auto rounded-md border'>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t('olt.pon')}</TableHead>
            <TableHead>{t('olt.onu_name')}</TableHead>
            <TableHead>{t('olt.mac')}</TableHead>
            <TableHead>{t('olt.state')}</TableHead>
            <TableHead>{t('olt.rx_power')}</TableHead>
            <TableHead>{t('olt.distance')}</TableHead>
            <TableHead>{t('olt.last_seen')}</TableHead>
            <TableHead className='text-right'>{t('common.action')}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {onus.map((onu) => {
            const adminEnabled = onu.adminState === 'ENABLE'
            return (
              <TableRow
                key={onu.id}
                className='cursor-pointer'
                onClick={() => onSelect(onu)}
              >
                <TableCell className='font-mono text-xs whitespace-nowrap'>
                  {onu.ponIndex}/{onu.onuIndex}
                </TableCell>
                <TableCell>
                  <div className='font-medium'>
                    {onu.name || (
                      <span className='text-muted-foreground'>
                        {positionLabel(onu.ponIndex, onu.onuIndex)}
                      </span>
                    )}
                  </div>
                  <div className='font-mono text-[11px] text-muted-foreground'>
                    {onu.serialNumber || '-'}
                  </div>
                </TableCell>
                <TableCell className='font-mono text-xs'>
                  {onu.macAddress || '-'}
                </TableCell>
                <TableCell>
                  <OnuStateBadge state={onu.state} />
                </TableCell>
                <TableCell>
                  <div className='flex flex-col gap-1'>
                    <span
                      className={`font-mono text-xs ${rxClassTextClass(onu.rxClass)}`}
                    >
                      {formatRxPower(onu.rxPowerDbm)}
                    </span>
                    <RxClassBadge rxClass={onu.rxClass} />
                  </div>
                </TableCell>
                <TableCell className='text-xs'>
                  {formatDistance(onu.distanceM)}
                </TableCell>
                <TableCell className='text-xs whitespace-nowrap text-muted-foreground'>
                  {formatDateTime(onu.lastSeenAt, language)}
                </TableCell>
                <TableCell
                  className='text-right'
                  onClick={(event) => event.stopPropagation()}
                >
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button
                        variant='ghost'
                        size='icon'
                        className='h-8 w-8 data-[state=open]:bg-muted'
                      >
                        <MoreHorizontal className='h-4 w-4' />
                        <span className='sr-only'>{t('common.action')}</span>
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align='end' className='w-48'>
                      <DropdownMenuLabel>
                        {positionLabel(onu.ponIndex, onu.onuIndex)}
                      </DropdownMenuLabel>
                      <DropdownMenuItem onClick={() => onRename(onu)}>
                        <Pencil className='me-2 h-3.5 w-3.5' />
                        {t('olt.rename')}
                      </DropdownMenuItem>
                      <DropdownMenuItem onClick={() => onReboot(onu)}>
                        <RotateCcw className='me-2 h-3.5 w-3.5 text-amber-500' />
                        {t('olt.reboot')}
                      </DropdownMenuItem>
                      <DropdownMenuItem onClick={() => onToggleAdmin(onu)}>
                        {adminEnabled ? (
                          <PowerOff className='me-2 h-3.5 w-3.5 text-rose-500' />
                        ) : (
                          <Power className='me-2 h-3.5 w-3.5 text-emerald-500' />
                        )}
                        {adminEnabled ? t('olt.disable') : t('olt.enable')}
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem
                        className='text-red-600 focus:text-red-600'
                        onClick={() => onDelete(onu)}
                      >
                        <Trash2 className='me-2 h-3.5 w-3.5' />
                        {t('olt.delete_onu')}
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </TableCell>
              </TableRow>
            )
          })}
        </TableBody>
      </Table>
    </div>
  )
}
