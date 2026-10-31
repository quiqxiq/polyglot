import { useState } from 'react'
import { Link } from '@tanstack/react-router'
import { CheckCircle2, Loader2, Receipt, UploadCloud } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import {
  useResolveManualPaymentMutation,
  useSubmitManualPaymentMutation,
} from '../api/use-payments'
import { formatIDR } from '../data/constants'

/**
 * Halaman publik pembayaran manual (/bayar): masukkan kode bayar, unggah
 * bukti transfer, lalu finance memverifikasi (planning v2 Fase 3).
 */
export function PayManualPage() {
  const [code, setCode] = useState('')
  const [invoice, setInvoice] = useState<
    | {
        invoiceId: string
        invoiceNumber: string
        period: string
        total: number
        paidAmount: number
        outstanding: number
        status: string
      }
    | null
  >(null)
  const [amount, setAmount] = useState('')
  const [senderName, setSenderName] = useState('')
  const [senderBank, setSenderBank] = useState('')
  const [notes, setNotes] = useState('')
  const [file, setFile] = useState<File | null>(null)
  const [uploading, setUploading] = useState(false)
  const [submitted, setSubmitted] = useState(false)

  const resolveMutation = useResolveManualPaymentMutation()
  const submitMutation = useSubmitManualPaymentMutation()

  const handleResolve = async () => {
    if (!code.trim()) {
      toast.error('Kode bayar wajib diisi')
      return
    }
    try {
      const res = await resolveMutation.mutateAsync(code.trim())
      setInvoice({
        invoiceId: res.invoiceId,
        invoiceNumber: res.invoiceNumber,
        period: res.period,
        total: res.total,
        paidAmount: res.paidAmount,
        outstanding: res.outstanding,
        status: res.status,
      })
      setAmount(String(res.outstanding || 0))
      if (res.outstanding <= 0) {
        toast.info('Tagihan ini sudah lunas.')
      }
    } catch {
      toast.error('Kode bayar tidak ditemukan atau tagihan tidak dapat dibayar.')
      setInvoice(null)
    }
  }

  const handleSubmit = async () => {
    if (!invoice) return
    if (!file) {
      toast.error('Unggah foto bukti transfer terlebih dahulu')
      return
    }
    if (!senderName.trim()) {
      toast.error('Nama pengirim wajib diisi')
      return
    }
    setUploading(true)
    try {
      const formData = new FormData()
      formData.append('file', file)
      const up = await fetch('/api/public/payment-proof', {
        method: 'POST',
        body: formData,
      })
      if (!up.ok) {
        const body = (await up.json().catch(() => ({}))) as { error?: string }
        throw new Error(body.error || 'Gagal mengunggah bukti')
      }
      const { url } = (await up.json()) as { url: string }

      await submitMutation.mutateAsync({
        manualPaymentCode: code.trim(),
        amount: Number(amount),
        senderName: senderName.trim(),
        senderBank: senderBank.trim(),
        proofImageUrl: url,
        notes: notes.trim(),
      })
      setSubmitted(true)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Gagal mengirim bukti')
    } finally {
      setUploading(false)
    }
  }

  if (submitted) {
    return (
      <div className='flex min-h-svh items-center justify-center bg-muted/30 p-4'>
        <Card className='w-full max-w-lg text-center'>
          <CardHeader>
            <div className='mx-auto flex h-14 w-14 items-center justify-center rounded-full bg-emerald-500/15 text-emerald-600'>
              <CheckCircle2 className='h-8 w-8' />
            </div>
            <CardTitle>Bukti Transfer Terkirim</CardTitle>
            <CardDescription>
              Bukti Anda sedang menunggu verifikasi finance. Layanan akan aktif
              kembali otomatis setelah disetujui.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button asChild variant='outline' className='w-full'>
              <Link to='/'>Kembali</Link>
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className='flex min-h-svh items-center justify-center bg-muted/30 p-4'>
      <Card className='w-full max-w-xl'>
        <CardHeader>
          <div className='flex items-center gap-2'>
            <Receipt className='h-5 w-5 text-primary' />
            <CardTitle>Bayar Tagihan (Upload Bukti Transfer)</CardTitle>
          </div>
          <CardDescription>
            Masukkan kode bayar pada tagihan/WhatsApp Anda, lalu unggah bukti
            transfer.
          </CardDescription>
        </CardHeader>
        <CardContent className='space-y-4'>
          <div className='flex gap-2'>
            <Input
              placeholder='Kode bayar (mis. PAY-123456)'
              value={code}
              onChange={(e) => setCode(e.target.value)}
              className='font-mono'
            />
            <Button onClick={handleResolve} disabled={resolveMutation.isPending}>
              {resolveMutation.isPending ? 'Mencari...' : 'Cari Tagihan'}
            </Button>
          </div>

          {invoice && (
            <div className='space-y-2 rounded-md border bg-muted/40 p-3 text-sm'>
              <div className='flex justify-between'>
                <span className='text-muted-foreground'>No. Invoice</span>
                <span className='font-mono font-semibold'>
                  {invoice.invoiceNumber}
                </span>
              </div>
              <div className='flex justify-between'>
                <span className='text-muted-foreground'>Periode</span>
                <span className='font-semibold'>{invoice.period}</span>
              </div>
              <div className='flex justify-between'>
                <span className='text-muted-foreground'>Total Tagihan</span>
                <span className='font-semibold'>{formatIDR(invoice.total)}</span>
              </div>
              <div className='flex justify-between border-t pt-2'>
                <span className='text-muted-foreground'>Sisa Tagihan</span>
                <span className='font-bold text-primary'>
                  {formatIDR(invoice.outstanding)}
                </span>
              </div>
            </div>
          )}

          {invoice && invoice.outstanding > 0 && (
            <div className='space-y-3'>
              <Input
                type='number'
                placeholder='Nominal transfer'
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
              />
              <div className='grid grid-cols-1 gap-3 sm:grid-cols-2'>
                <Input
                  placeholder='Nama pengirim'
                  value={senderName}
                  onChange={(e) => setSenderName(e.target.value)}
                />
                <Input
                  placeholder='Bank pengirim'
                  value={senderBank}
                  onChange={(e) => setSenderBank(e.target.value)}
                />
              </div>
              <label className='flex cursor-pointer items-center gap-2 rounded-md border border-dashed p-3 text-sm'>
                <UploadCloud className='h-5 w-5 text-muted-foreground' />
                <span className='text-muted-foreground'>
                  {file ? file.name : 'Pilih foto bukti transfer (JPG/PNG/WebP, maks 5MB)'}
                </span>
                <input
                  type='file'
                  accept='image/jpeg,image/png,image/webp'
                  className='hidden'
                  onChange={(e) => setFile(e.target.files?.[0] || null)}
                />
              </label>
              <Textarea
                placeholder='Catatan (opsional)'
                rows={2}
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
              />
              <Button
                className='w-full'
                onClick={handleSubmit}
                disabled={uploading || submitMutation.isPending}
              >
                {uploading || submitMutation.isPending ? (
                  <>
                    <Loader2 className='mr-2 h-4 w-4 animate-spin' /> Mengirim...
                  </>
                ) : (
                  'Kirim Bukti Transfer'
                )}
              </Button>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
