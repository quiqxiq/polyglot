import { useState } from 'react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { toast } from 'sonner'
import { useCreateWASessionMutation } from '../api/use-whatsapp'
import { QRModal } from './qr-modal'
import type { WASession } from '@/gen/v1/whatsapp_pb'

interface CreateDeviceDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function CreateDeviceDialog({ open, onOpenChange }: CreateDeviceDialogProps) {
  const [name, setName] = useState('')
  const [phoneNumber, setPhoneNumber] = useState('')
  const [qrSession, setQrSession] = useState<WASession | null>(null)

  const createMutation = useCreateWASessionMutation()

  // Reset form saat dialog ditutup (batal / sukses), bukan via effect.
  const handleOpenChange = (next: boolean) => {
    if (!next) {
      setName('')
      setPhoneNumber('')
    }
    onOpenChange(next)
  }

  const handleCreate = () => {
    if (!name.trim()) {
      toast.error('Nama perangkat wajib diisi')
      return
    }
    createMutation.mutate(
      { name: name.trim(), phoneNumber: phoneNumber.trim() },
      {
        onSuccess: (res) => {
          toast.success('Perangkat dibuat — scan QR untuk menghubungkan')
          handleOpenChange(false)
          if (res.session) {
            setQrSession(res.session)
          }
        },
        onError: (err) => toast.error(`Gagal membuat perangkat: ${err.message}`),
      },
    )
  }

  return (
    <>
      <Dialog open={open} onOpenChange={handleOpenChange}>
        <DialogContent className='sm:max-w-md'>
          <DialogHeader>
            <DialogTitle>Tambah Perangkat WhatsApp</DialogTitle>
            <DialogDescription>
              Buat slot perangkat baru. Setelah dibuat, QR akan langsung muncul untuk di-scan
              dari WhatsApp di HP Anda.
            </DialogDescription>
          </DialogHeader>

          <div className='space-y-4'>
            <div className='space-y-2'>
              <Label htmlFor='wa-device-name'>Nama perangkat</Label>
              <Input
                id='wa-device-name'
                placeholder='cth: HP Support Toko'
                value={name}
                onChange={(e) => setName(e.target.value)}
                autoFocus
              />
            </div>
            <div className='space-y-2'>
              <Label htmlFor='wa-device-phone'>Nomor WhatsApp (opsional)</Label>
              <Input
                id='wa-device-phone'
                placeholder='cth: 6281234567890'
                value={phoneNumber}
                onChange={(e) => setPhoneNumber(e.target.value)}
              />
              <p className='text-xs text-muted-foreground'>
                Isi untuk pairing code. Kosongkan bila hanya ingin scan QR.
              </p>
            </div>
          </div>

          <DialogFooter>
            <Button variant='outline' onClick={() => handleOpenChange(false)} disabled={createMutation.isPending}>
              Batal
            </Button>
            <Button onClick={handleCreate} disabled={createMutation.isPending} className='gap-1.5'>
              {createMutation.isPending ? 'Membuat…' : 'Buat & Tampilkan QR'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <QRModal key={qrSession?.id ?? 'closed'} session={qrSession} onOpenChange={(open) => !open && setQrSession(null)} />
    </>
  )
}
