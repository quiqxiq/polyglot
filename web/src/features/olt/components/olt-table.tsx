import type { OltDevice } from '@/gen/v1/olt_pb'
import {
  Eye,
  MoreHorizontal,
  Pencil,
  RefreshCw,
  Trash2,
  Zap,
} from 'lucide-react'
import { useI18n } from '@/context/i18n-provider'
import { Badge } from '@/components/ui/badge'
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
import { EnabledBadge } from './olt-badges'

interface OltTableProps {
  olts: OltDevice[]
  isLoading: boolean
  onDetail: (olt: OltDevice) => void
  onEdit: (olt: OltDevice) => void
  onDelete: (olt: OltDevice) => void
  onTest: (olt: OltDevice) => void
  onSync: (olt: OltDevice) => void
}

export function OltTable({
  olts,
  isLoading,
  onDetail,
  onEdit,
  onDelete,
  onTest,
  onSync,
}: OltTableProps) {
  const { t } = useI18n()

  if (isLoading) {
    return (
      <div className='space-y-2 rounded-md border p-4'>
        {Array.from({ length: 4 }).map((_, index) => (
          <Skeleton key={index} className='h-10 w-full' />
        ))}
      </div>
    )
  }

  if (olts.length === 0) {
    return (
      <div className='rounded-md border py-12 text-center text-sm text-muted-foreground'>
        {t('olt.empty')}
      </div>
    )
  }

  return (
    <div className='overflow-x-auto rounded-md border'>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t('olt.name')}</TableHead>
            <TableHead>{t('olt.host')}</TableHead>
            <TableHead>{t('olt.variant')}</TableHead>
            <TableHead>{t('olt.transport')}</TableHead>
            <TableHead>{t('olt.onu_online_total')}</TableHead>
            <TableHead>{t('olt.status')}</TableHead>
            <TableHead className='text-right'>{t('common.action')}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {olts.map((olt) => (
            <TableRow key={olt.id}>
              <TableCell>
                <button
                  type='button'
                  className='text-start font-semibold text-primary hover:underline'
                  onClick={() => onDetail(olt)}
                >
                  {olt.name || olt.host}
                </button>
                <div className='text-xs text-muted-foreground'>
                  {t('olt.pon_count')}: {olt.ponCount}
                </div>
              </TableCell>
              <TableCell className='font-mono text-xs'>{olt.host}</TableCell>
              <TableCell>
                <Badge variant='outline'>{olt.variant || 'AUTO'}</Badge>
              </TableCell>
              <TableCell className='text-xs uppercase'>
                {olt.transport}
              </TableCell>
              <TableCell>
                <span className='font-mono text-sm'>
                  {olt.onlineOnus}/{olt.totalOnus}
                </span>
                <div className='text-xs text-muted-foreground'>
                  {t('olt.offline_onus')}: {olt.offlineOnus}
                </div>
              </TableCell>
              <TableCell>
                <EnabledBadge enabled={olt.enabled} />
              </TableCell>
              <TableCell className='text-right'>
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
                      {olt.name || olt.host}
                    </DropdownMenuLabel>
                    <DropdownMenuItem onClick={() => onDetail(olt)}>
                      <Eye className='me-2 h-3.5 w-3.5' />
                      {t('common.details')}
                    </DropdownMenuItem>
                    <DropdownMenuItem onClick={() => onEdit(olt)}>
                      <Pencil className='me-2 h-3.5 w-3.5' />
                      {t('olt.edit')}
                    </DropdownMenuItem>
                    <DropdownMenuItem onClick={() => onTest(olt)}>
                      <Zap className='me-2 h-3.5 w-3.5 text-amber-500' />
                      {t('olt.test')}
                    </DropdownMenuItem>
                    <DropdownMenuItem onClick={() => onSync(olt)}>
                      <RefreshCw className='me-2 h-3.5 w-3.5 text-emerald-500' />
                      {t('olt.sync')}
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem
                      className='text-red-600 focus:text-red-600'
                      onClick={() => onDelete(olt)}
                    >
                      <Trash2 className='me-2 h-3.5 w-3.5' />
                      {t('olt.delete')}
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
