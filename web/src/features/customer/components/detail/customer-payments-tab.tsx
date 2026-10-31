import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

interface PaymentRow {
  id: string
  paymentNo: string
  invoiceId: string
  amount: number
  paymentDateUnix: bigint | number
  scanMethod: string
  voided: boolean
  voidReason: string
}

interface CustomerPaymentsTabProps {
  payments: PaymentRow[]
  isLoading?: boolean
}

function formatDate(unix: bigint | number | undefined) {
  const n = Number(unix || 0)
  if (!n) return '-'
  return new Date(n * 1000).toLocaleString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function formatIDR(v: number | undefined) {
  return `Rp${Number(v || 0).toLocaleString('id-ID')}`
}

/** Tab riwayat pembayaran pelanggan (Customer 360 — Fase 5). */
export function CustomerPaymentsTab({ payments, isLoading }: CustomerPaymentsTabProps) {
  if (isLoading) {
    return <p className='py-8 text-center text-sm text-muted-foreground'>Memuat...</p>
  }
  if (payments.length === 0) {
    return (
      <Card className='border-dashed shadow-none'>
        <CardContent className='py-10 text-center text-sm text-muted-foreground'>
          Belum ada pembayaran tercatat.
        </CardContent>
      </Card>
    )
  }
  return (
    <div className='rounded-md border'>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>No. Kwitansi</TableHead>
            <TableHead>Tanggal</TableHead>
            <TableHead>Invoice</TableHead>
            <TableHead>Metode</TableHead>
            <TableHead className='text-right'>Nominal</TableHead>
            <TableHead>Status</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {payments.map((p) => (
            <TableRow key={p.id}>
              <TableCell className='font-mono text-xs'>{p.paymentNo}</TableCell>
              <TableCell className='text-xs'>{formatDate(p.paymentDateUnix)}</TableCell>
              <TableCell className='font-mono text-xs'>{p.invoiceId}</TableCell>
              <TableCell className='text-xs'>{p.scanMethod || '-'}</TableCell>
              <TableCell className='text-right text-xs font-semibold'>
                {formatIDR(p.amount)}
              </TableCell>
              <TableCell>
                {p.voided ? (
                  <Badge
                    variant='outline'
                    className='bg-rose-500/15 text-rose-600 border-rose-500/30'
                    title={p.voidReason}
                  >
                    Void
                  </Badge>
                ) : (
                  <Badge
                    variant='outline'
                    className='bg-emerald-500/15 text-emerald-600 border-emerald-500/30'
                  >
                    Sah
                  </Badge>
                )}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
