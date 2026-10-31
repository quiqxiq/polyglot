import { CheckCircle2, ExternalLink, XCircle } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import type { ManualPaymentSubmission } from '@/gen/v1/billing_pb'
import { formatIDR, submissionStatusBadge, SUBMISSION_STATUS } from '../data/constants'

interface PaymentsTableProps {
  data: ManualPaymentSubmission[]
  isLoading?: boolean
  onApprove: (submission: ManualPaymentSubmission) => void
  onReject: (submission: ManualPaymentSubmission) => void
}

function formatUnixDate(unix: bigint | number | undefined) {
  const num = Number(unix || 0)
  if (!num) return '-'
  return new Date(num * 1000).toLocaleString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function PaymentsTable({
  data,
  isLoading,
  onApprove,
  onReject,
}: PaymentsTableProps) {
  if (isLoading) {
    return (
      <p className='py-8 text-center text-sm text-muted-foreground'>Memuat...</p>
    )
  }
  if (data.length === 0) {
    return (
      <p className='py-8 text-center text-sm text-muted-foreground'>
        Belum ada bukti transfer.
      </p>
    )
  }
  return (
    <div className='rounded-md border'>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Dibuat</TableHead>
            <TableHead>Invoice</TableHead>
            <TableHead>Pengirim</TableHead>
            <TableHead className='text-right'>Nominal</TableHead>
            <TableHead>Bukti</TableHead>
            <TableHead>Status</TableHead>
            <TableHead className='text-right'>Aksi</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((sub) => {
            const meta = submissionStatusBadge(sub.status)
            const isPending = sub.status === SUBMISSION_STATUS.PENDING
            return (
              <TableRow key={sub.id}>
                <TableCell className='whitespace-nowrap text-xs'>
                  {formatUnixDate(sub.createdAtUnix)}
                </TableCell>
                <TableCell className='font-mono text-xs'>{sub.invoiceId}</TableCell>
                <TableCell className='text-xs'>
                  <div className='font-semibold'>{sub.senderName || '-'}</div>
                  <div className='text-muted-foreground'>{sub.senderBank || '-'}</div>
                </TableCell>
                <TableCell className='text-right font-semibold'>
                  {formatIDR(sub.amount)}
                </TableCell>
                <TableCell>
                  {sub.proofImageUrl ? (
                    <a
                      href={sub.proofImageUrl}
                      target='_blank'
                      rel='noopener noreferrer'
                      className='inline-flex items-center gap-1 text-xs text-primary underline'
                    >
                      <ExternalLink className='h-3.5 w-3.5' /> Lihat
                    </a>
                  ) : (
                    '-'
                  )}
                </TableCell>
                <TableCell>
                  <Badge variant='outline' className={meta.className}>
                    {meta.label}
                  </Badge>
                  {sub.rejectionReason && (
                    <p className='mt-1 max-w-[12rem] text-[10px] text-rose-600'>
                      {sub.rejectionReason}
                    </p>
                  )}
                </TableCell>
                <TableCell className='text-right'>
                  {isPending ? (
                    <div className='flex justify-end gap-1.5'>
                      <Button
                        size='sm'
                        variant='outline'
                        className='h-7 gap-1 border-emerald-500/40 text-emerald-600'
                        onClick={() => onApprove(sub)}
                      >
                        <CheckCircle2 className='h-3.5 w-3.5' /> Setujui
                      </Button>
                      <Button
                        size='sm'
                        variant='outline'
                        className='h-7 gap-1 border-rose-500/40 text-rose-600'
                        onClick={() => onReject(sub)}
                      >
                        <XCircle className='h-3.5 w-3.5' /> Tolak
                      </Button>
                    </div>
                  ) : (
                    <span className='text-[10px] text-muted-foreground'>-</span>
                  )}
                </TableCell>
              </TableRow>
            )
          })}
        </TableBody>
      </Table>
    </div>
  )
}
