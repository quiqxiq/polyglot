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

interface PSBRow {
  id: string
  registrationNo: string
  status: string
  createdAtUnix: bigint | number
  installedAtUnix: bigint | number
  surveyStatus: string
}

interface CustomerPSBTabProps {
  registrations: PSBRow[]
  isLoading?: boolean
}

function formatDate(unix: bigint | number | undefined) {
  const n = Number(unix || 0)
  if (!n) return '-'
  return new Date(n * 1000).toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}

const STATUS_CLASS: Record<string, string> = {
  ACTIVE: 'bg-emerald-500/15 text-emerald-600 border-emerald-500/30',
  INSTALLED: 'bg-purple-500/15 text-purple-600 border-purple-500/30',
  ON_PROGRESS: 'bg-indigo-500/15 text-indigo-600 border-indigo-500/30',
  SURVEY: 'bg-cyan-500/15 text-cyan-600 border-cyan-500/30',
  APPROVED: 'bg-blue-500/15 text-blue-600 border-blue-500/30',
  PENDING: 'bg-amber-500/15 text-amber-600 border-amber-500/30',
  REJECTED: 'bg-rose-500/15 text-rose-600 border-rose-500/30',
  CANCELLED: 'bg-zinc-500/15 text-zinc-600 border-zinc-500/30',
}

/** Tab riwayat pendaftaran PSB pelanggan (Customer 360 — Fase 5). */
export function CustomerPSBTab({ registrations, isLoading }: CustomerPSBTabProps) {
  if (isLoading) {
    return <p className='py-8 text-center text-sm text-muted-foreground'>Memuat...</p>
  }
  if (registrations.length === 0) {
    return (
      <Card className='border-dashed shadow-none'>
        <CardContent className='py-10 text-center text-sm text-muted-foreground'>
          Belum ada riwayat pendaftaran PSB untuk pelanggan ini.
        </CardContent>
      </Card>
    )
  }
  return (
    <div className='rounded-md border'>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>No. Registrasi</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Survey</TableHead>
            <TableHead>Tanggal Daftar</TableHead>
            <TableHead>Tanggal Pasang</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {registrations.map((r) => (
            <TableRow key={r.id}>
              <TableCell className='font-mono text-xs'>{r.registrationNo}</TableCell>
              <TableCell>
                <Badge
                  variant='outline'
                  className={STATUS_CLASS[r.status] || 'bg-muted text-muted-foreground'}
                >
                  {r.status}
                </Badge>
              </TableCell>
              <TableCell className='text-xs'>{r.surveyStatus || '-'}</TableCell>
              <TableCell className='text-xs'>{formatDate(r.createdAtUnix)}</TableCell>
              <TableCell className='text-xs'>{formatDate(r.installedAtUnix)}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
