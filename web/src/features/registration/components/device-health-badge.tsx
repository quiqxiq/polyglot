import { useEffect, useState } from 'react'
import { Activity, AlertTriangle, CheckCircle2, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useTestConnectionMutation } from '@/features/devices/api/use-devices'
import { TestDeviceConnectionRequest } from '@/gen/v1/device_pb'

interface DeviceHealthBadgeProps {
  deviceId: string
}

/**
 * Verifikasi kesehatan router (ping/API status) sebelum memilih BRAS pada
 * form pasang baru atau aktivasi — planning v2 Fase 0.
 */
export function DeviceHealthBadge({ deviceId }: DeviceHealthBadgeProps) {
  const testConnection = useTestConnectionMutation()
  const [health, setHealth] = useState<
    { ok: boolean; latencyMs: number; message: string } | null
  >(null)
  const [checking, setChecking] = useState(false)

  const runCheck = async (id: string) => {
    if (!id) return
    setChecking(true)
    setHealth(null)
    try {
      const res = await testConnection.mutateAsync(
        new TestDeviceConnectionRequest({ id })
      )
      setHealth({
        ok: Boolean(res.success),
        latencyMs: Number(res.latencyMs || 0),
        message: res.message || res.status || '',
      })
    } catch (err) {
      setHealth({
        ok: false,
        latencyMs: 0,
        message: err instanceof Error ? err.message : 'Router tidak terjangkau',
      })
    } finally {
      setChecking(false)
    }
  }

  useEffect(() => {
    if (deviceId) void runCheck(deviceId)
    else setHealth(null)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [deviceId])

  if (!deviceId) {
    return (
      <p className='flex items-center gap-1.5 text-xs text-muted-foreground'>
        <Activity className='h-3.5 w-3.5' />
        Pilih router untuk memeriksa kesehatan koneksi.
      </p>
    )
  }

  return (
    <div className='flex items-center justify-between gap-2 rounded-md border border-dashed px-2.5 py-1.5 text-xs'>
      <span className='flex items-center gap-1.5'>
        {checking ? (
          <>
            <Loader2 className='h-3.5 w-3.5 animate-spin text-muted-foreground' />
            <span className='text-muted-foreground'>Memeriksa router...</span>
          </>
        ) : health ? (
          health.ok ? (
            <>
              <CheckCircle2 className='h-3.5 w-3.5 text-emerald-600' />
              <span className='text-emerald-700 dark:text-emerald-400'>
                Router terhubung · {health.latencyMs} ms
              </span>
            </>
          ) : (
            <>
              <AlertTriangle className='h-3.5 w-3.5 text-rose-600' />
              <span className='text-rose-700 dark:text-rose-400'>
                Router tidak terjangkau{health.message ? `: ${health.message}` : ''}
              </span>
            </>
          )
        ) : (
          <span className='text-muted-foreground'>Belum diperiksa</span>
        )}
      </span>
      <Button
        type='button'
        size='sm'
        variant='ghost'
        className='h-6 px-2 text-[11px]'
        onClick={() => runCheck(deviceId)}
        disabled={checking}
      >
        Cek Ulang
      </Button>
    </div>
  )
}
