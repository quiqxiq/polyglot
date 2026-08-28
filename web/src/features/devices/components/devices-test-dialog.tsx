'use client'

import { useState } from 'react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'
import { useDevicesContext } from './devices-provider'
import { useTestConnectionMutation } from '../api/use-devices'
import {
  TestDeviceConnectionRequest,
  type TestDeviceConnectionResponse,
} from '@/gen/v1/device_pb'
import { getErrorMessage } from '../lib/formatters'
import { Loader2, Zap } from 'lucide-react'

export function DevicesTestDialog() {
  const { open, setOpen, currentRow, setCurrentRow } = useDevicesContext()
  const testMutation = useTestConnectionMutation()
  const [result, setResult] = useState<TestDeviceConnectionResponse | null>(null)

  async function handleTest() {
    if (!currentRow) return
    setResult(null)

    try {
      const res = await testMutation.mutateAsync(
        new TestDeviceConnectionRequest({ id: currentRow.id })
      )
      setResult(res)
      if (res.success) {
        toast.success(`Koneksi ke ${currentRow.name} berhasil!`)
      } else {
        toast.error(`Koneksi gagal: ${res.message}`)
      }
    } catch (err: unknown) {
      toast.error(getErrorMessage(err, 'Gagal menguji koneksi perangkat'))
    }
  }

  return (
    <Dialog
      open={open === 'test'}
      onOpenChange={() => {
        setOpen(null)
        setCurrentRow(null)
        setResult(null)
      }}
    >
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <DialogTitle className='flex items-center gap-2'>
            <Zap className='h-5 w-5 text-amber-500' />
            Test Device Connection
          </DialogTitle>
          <DialogDescription>
            Uji konektivitas realtime ke <strong>{currentRow?.name}</strong> (
            {currentRow?.host}:{currentRow?.port}).
          </DialogDescription>
        </DialogHeader>

        <div className='space-y-4 py-2'>
          <div className='rounded-lg border p-3 text-sm space-y-1.5 bg-muted/40'>
            <div className='flex justify-between'>
              <span className='text-muted-foreground'>Host:</span>
              <span className='font-mono font-medium'>
                {currentRow?.host}:{currentRow?.port}
              </span>
            </div>
            <div className='flex justify-between'>
              <span className='text-muted-foreground'>Vendor / Driver:</span>
              <span className='font-medium'>
                {currentRow?.vendor} ({currentRow?.driverType})
              </span>
            </div>
          </div>

          {testMutation.isPending && (
            <div className='flex flex-col items-center justify-center py-6 space-y-2'>
              <Loader2 className='h-8 w-8 animate-spin text-primary' />
              <p className='text-xs text-muted-foreground'>Menghubungkan ke perangkat...</p>
            </div>
          )}

          {result && (
            <div className='rounded-lg border p-4 space-y-2 text-sm'>
              <div className='flex items-center justify-between'>
                <span className='font-medium'>Status Hasil:</span>
                <Badge variant={result.success ? 'default' : 'destructive'}>
                  {result.status.toUpperCase()}
                </Badge>
              </div>

              {result.latencyMs > 0 && (
                <div className='flex justify-between text-xs'>
                  <span className='text-muted-foreground'>Latency:</span>
                  <span className='font-mono'>{result.latencyMs} ms</span>
                </div>
              )}

              {result.version && (
                <div className='flex justify-between text-xs'>
                  <span className='text-muted-foreground'>OS Version:</span>
                  <span className='font-mono'>{result.version}</span>
                </div>
              )}

              {result.uptime && (
                <div className='flex justify-between text-xs'>
                  <span className='text-muted-foreground'>Uptime:</span>
                  <span>{result.uptime}</span>
                </div>
              )}

              {result.boardName && (
                <div className='flex justify-between text-xs'>
                  <span className='text-muted-foreground'>Board Name:</span>
                  <span>{result.boardName}</span>
                </div>
              )}

              {result.message && (
                <div className='mt-2 pt-2 border-t text-xs text-muted-foreground'>
                  <p className='font-medium text-foreground mb-0.5'>Pesan Respons:</p>
                  <p className='font-mono bg-muted p-2 rounded text-[11px]'>{result.message}</p>
                </div>
              )}
            </div>
          )}
        </div>

        <DialogFooter>
          <Button
            type='button'
            variant='outline'
            onClick={() => {
              setOpen(null)
              setCurrentRow(null)
              setResult(null)
            }}
          >
            Tutup
          </Button>
          <Button onClick={handleTest} disabled={testMutation.isPending}>
            {testMutation.isPending ? 'Menguji...' : 'Jalankan Uji Sekarang'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
