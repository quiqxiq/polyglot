import { useState } from 'react'
import { CheckCircle2, Copy, LoaderCircle, QrCode, RefreshCw, Smartphone } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { toast } from 'sonner'
import { useWASessionQRQuery, useGetPairingCodeMutation } from '../api/use-whatsapp'
import type { WASession } from '@/gen/v1/whatsapp_pb'

interface QRModalProps {
  session: WASession | null
  onOpenChange: (open: boolean) => void
}

export function QRModal({ session, onOpenChange }: QRModalProps) {
  // State pairPhone/pairCode di-reset otomatis oleh React saat modal di-remount
  // via prop `key` dari parent (lihat penggunaan <QRModal key={session?.id} />).
  const [pairPhone, setPairPhone] = useState('')
  const [pairCode, setPairCode] = useState('')

  const sessionId = session?.id ?? ''
  const open = Boolean(session)
  // Device yang sudah online tidak perlu QR — tampilkan state terhubung dan
  // hentikan polling QR (session.status adalah snapshot saat modal dibuka).
  const isOnline = session?.status === 'online'

  const qrQuery = useWASessionQRQuery(sessionId, open && !isOnline)
  const pairingMutation = useGetPairingCodeMutation()

  const qrBase64 = qrQuery.data?.qrCodeBase64 ?? ''
  const hasQr = Boolean(qrBase64)

  const handleRequestPairing = () => {
    if (!pairPhone.trim()) {
      toast.error('Masukkan nomor WhatsApp tujuan pairing')
      return
    }
    pairingMutation.mutate(
      { sessionId, phoneNumber: pairPhone.trim() },
      {
        onSuccess: (res) => {
          setPairCode(res.pairingCode || '')
          if (res.pairingCode) {
            toast.success('Kode pairing berhasil dibuat')
          }
        },
        onError: (err) => toast.error(`Gagal membuat kode pairing: ${err.message}`),
      },
    )
  }

  const handleCopy = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text)
      toast.success('Disalin ke clipboard')
    } catch {
      toast.error('Gagal menyalin')
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <DialogTitle className='flex items-center gap-2'>
            <QrCode className='size-5 text-primary' />
            {session?.name || 'Perangkat'} — Hubungkan
          </DialogTitle>
          <DialogDescription>
            Buka WhatsApp di HP → Setelan → Perangkat tertaut → Tautkan perangkat.
          </DialogDescription>
        </DialogHeader>

        {isOnline ? (
          <div className='flex flex-col items-center gap-3 py-6 text-center'>
            <div className='flex size-16 items-center justify-center rounded-full bg-emerald-500/15'>
              <CheckCircle2 className='size-8 text-emerald-600' />
            </div>
            <div>
              <p className='font-semibold'>Perangkat sudah terhubung</p>
              <p className='text-sm text-muted-foreground'>
                {session?.name} online dan siap menerima pesan.
              </p>
            </div>
          </div>
        ) : (
        <Tabs defaultValue='qr'>
          <TabsList className='grid w-full grid-cols-2'>
            <TabsTrigger value='qr'>Scan QR</TabsTrigger>
            <TabsTrigger value='pairing'>Pairing Code</TabsTrigger>
          </TabsList>

          {/* ── Tab QR ─────────────────────────────────────────── */}
          <TabsContent value='qr' className='flex flex-col items-center gap-4 py-4'>
            {qrQuery.isFetching && !hasQr ? (
              <div className='flex size-56 items-center justify-center rounded-xl border bg-muted/40'>
                <LoaderCircle className='size-8 animate-spin text-muted-foreground' />
              </div>
            ) : hasQr ? (
              <img
                src={qrBase64}
                alt='WhatsApp QR'
                className='size-56 rounded-xl border bg-white p-2'
              />
            ) : (
              <div className='flex size-56 flex-col items-center justify-center gap-2 rounded-xl border bg-muted/40 text-center text-sm text-muted-foreground'>
                <Smartphone className='size-8' />
                <span>Menunggu QR…</span>
                <span className='max-w-40 text-xs'>
                  Jika lama, coba tombol Muat Ulang QR.
                </span>
              </div>
            )}
            <Button
              variant='outline'
              size='sm'
              className='gap-1.5'
              disabled={qrQuery.isFetching}
              onClick={() => qrQuery.refetch()}
            >
              <RefreshCw size={14} /> Muat Ulang QR
            </Button>
            <p className='max-w-72 text-center text-xs text-muted-foreground'>
              QR berlaku ±60 detik dan diperbarui otomatis secara live sampai
              perangkat berhasil ditautkan.
            </p>
          </TabsContent>

          {/* ── Tab Pairing Code ───────────────────────────────── */}
          <TabsContent value='pairing' className='space-y-4 py-4'>
            {pairCode ? (
              <div className='flex flex-col items-center gap-3'>
                <div className='rounded-xl border bg-muted/40 px-8 py-4 text-center'>
                  <p className='text-xs uppercase tracking-wide text-muted-foreground'>
                    Masukkan kode ini di HP
                  </p>
                  <p className='mt-2 font-mono text-2xl font-bold tracking-widest'>
                    {pairCode}
                  </p>
                </div>
                <Button size='sm' variant='outline' className='gap-1.5' onClick={() => handleCopy(pairCode)}>
                  <Copy size={14} /> Salin Kode
                </Button>
              </div>
            ) : (
              <>
                <div className='space-y-2'>
                  <Label htmlFor='pair-phone'>Nomor WhatsApp tujuan pairing</Label>
                  <Input
                    id='pair-phone'
                    placeholder='cth: 6281234567890'
                    value={pairPhone}
                    onChange={(e) => setPairPhone(e.target.value)}
                  />
                  <p className='text-xs text-muted-foreground'>
                    Kode 8 digit akan tampil setelah diminta; masukkan di HP di menu
                    'Tautkan perangkat' → 'Tautkan dengan nomor telepon'.
                  </p>
                </div>
                <Button
                  className='w-full gap-1.5'
                  onClick={handleRequestPairing}
                  disabled={pairingMutation.isPending}
                >
                  {pairingMutation.isPending ? 'Meminta kode…' : 'Minta Kode Pairing'}
                </Button>
              </>
            )}
          </TabsContent>
        </Tabs>
        )}
      </DialogContent>
    </Dialog>
  )
}
