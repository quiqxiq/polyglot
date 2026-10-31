import { useState } from 'react'
import { Link, useSearch } from '@tanstack/react-router'
import { History, Loader2, Search } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { useTrackRegistrationQuery } from '../api/use-registration'
import { registrationStatusBadge } from '../data/constants'

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

function TrackResult({ token }: { token: string }) {
  const { data, isLoading, isError } = useTrackRegistrationQuery(token)
  if (isLoading) {
    return (
      <div className='flex justify-center py-8'>
        <Loader2 className='h-6 w-6 animate-spin text-muted-foreground' />
      </div>
    )
  }
  if (isError || !data?.registration) {
    return (
      <p className='py-6 text-center text-sm text-muted-foreground'>
        Pendaftaran dengan token tersebut tidak ditemukan.
      </p>
    )
  }
  const reg = data.registration
  const statusMeta = registrationStatusBadge(reg.status)
  return (
    <div className='space-y-4'>
      <div className='flex items-start justify-between gap-3'>
        <div>
          <p className='font-semibold'>{reg.fullName}</p>
          <p className='font-mono text-xs text-muted-foreground'>
            {reg.registrationNo}
          </p>
        </div>
        <Badge variant='outline' className={statusMeta.className}>
          {statusMeta.label}
        </Badge>
      </div>

      {reg.surveyStatus && (
        <div className='rounded-md bg-muted p-3 text-xs'>
          <p>
            <span className='text-muted-foreground'>Hasil survey:</span>{' '}
            <span className='font-semibold'>{reg.surveyStatus}</span>
            {reg.estimatedCableLengthMeter
              ? ` — estimasi kabel ${reg.estimatedCableLengthMeter} m`
              : ''}
          </p>
          {reg.surveyNotes && (
            <p className='mt-1 text-muted-foreground'>{reg.surveyNotes}</p>
          )}
        </div>
      )}

      {reg.scheduledInstallDateUnix ? (
        <div className='rounded-md bg-muted p-3 text-xs'>
          <p>
            <span className='text-muted-foreground'>Jadwal pemasangan:</span>{' '}
            <span className='font-semibold'>
              {formatUnixDate(reg.scheduledInstallDateUnix)}
            </span>
            {reg.scheduledInstallTime ? ` (${reg.scheduledInstallTime})` : ''}
          </p>
        </div>
      ) : null}

      <div>
        <h4 className='mb-2 flex items-center gap-2 text-sm font-semibold'>
          <History className='h-4 w-4 text-muted-foreground' />
          Riwayat Status
        </h4>
        {data.events.length === 0 ? (
          <p className='text-xs text-muted-foreground'>Belum ada event.</p>
        ) : (
          <ol className='relative space-y-3 border-l pl-4'>
            {data.events.map((ev) => (
              <li key={ev.id} className='text-xs'>
                <div className='flex items-center justify-between gap-2'>
                  <span className='font-semibold'>{ev.title || ev.eventType}</span>
                  <span className='text-[10px] text-muted-foreground'>
                    {formatUnixDate(ev.createdAtUnix)}
                  </span>
                </div>
                {ev.description && (
                  <p className='text-muted-foreground'>{ev.description}</p>
                )}
              </li>
            ))}
          </ol>
        )}
      </div>
    </div>
  )
}

/**
 * Halaman publik lacak status pendaftaran (/register/status?token=...) —
 * tanpa login, hanya menampilkan data pendaftaran milik token tersebut.
 */
export function TrackRegistrationPage() {
  const search = useSearch({ from: '/register/status' })
  const initial = (search.token as string) || ''
  const [tokenInput, setTokenInput] = useState(initial)
  const [activeToken, setActiveToken] = useState(initial)

  return (
    <div className='flex min-h-svh items-center justify-center bg-muted/30 p-4'>
      <Card className='w-full max-w-xl'>
        <CardHeader>
          <CardTitle>Lacak Status Pendaftaran</CardTitle>
          <CardDescription>
            Masukkan token lacak yang Anda terima setelah mendaftar.
          </CardDescription>
        </CardHeader>
        <CardContent className='space-y-4'>
          <div className='flex gap-2'>
            <Input
              placeholder='Token lacak (32 karakter)'
              value={tokenInput}
              onChange={(e) => setTokenInput(e.target.value)}
              className='font-mono'
            />
            <Button onClick={() => setActiveToken(tokenInput.trim())}>
              <Search className='mr-1.5 h-4 w-4' /> Lacak
            </Button>
          </div>
          {activeToken ? <TrackResult token={activeToken} /> : null}
          <p className='text-center text-xs text-muted-foreground'>
            Belum mendaftar?{' '}
            <Link to='/register' className='text-primary underline'>
              Daftar sekarang
            </Link>
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
