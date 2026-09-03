import { useState } from 'react'
import { toast } from 'sonner'
import {
  Building2,
  Check,
  Copy,
  CreditCard,
  ExternalLink,
  QrCode,
  Sparkles,
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { type Invoice } from '@/gen/v1/billing_pb'
import { useCashierChargeMutation, type CashierChargeResponse } from '../api/use-invoices'

interface CashierOnlineChargeProps {
  invoice: Invoice
}

function formatCurrency(amount: number) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(amount)
}

const CHANNELS = [
  { value: 'QRIS', label: 'QRIS (BCA, Livin, GoPay, OVO, ShopeePay, DANA)', icon: QrCode },
  { value: 'BRIVA', label: 'BRI Virtual Account (BRIVA)', icon: Building2 },
  { value: 'BNIVA', label: 'BNI Virtual Account', icon: Building2 },
  { value: 'MANDIRIVA', label: 'Mandiri Virtual Account', icon: Building2 },
  { value: 'BCAVA', label: 'BCA Virtual Account', icon: Building2 },
]

export function CashierOnlineCharge({ invoice }: CashierOnlineChargeProps) {
  const [channel, setChannel] = useState('QRIS')
  const [copied, setCopied] = useState(false)
  const [chargeResult, setChargeResult] = useState<CashierChargeResponse | null>(null)

  const chargeMutation = useCashierChargeMutation()
  const outstanding = Math.max(0, invoice.total - invoice.paidAmount)

  const handleGenerate = async () => {
    try {
      const res = await chargeMutation.mutateAsync({
        invoice_id: invoice.id,
        channel,
        expire_minutes: 60,
      })
      setChargeResult(res)
      toast.success('Tagihan online Tripay berhasil dibuat')
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Gagal membuat tagihan online'
      toast.error('Gagal membuat tagihan online', { description: message })
    }
  }

  const handleCopy = (text: string, label: string) => {
    navigator.clipboard.writeText(text)
    setCopied(true)
    toast.success(`${label} disalin`)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className='space-y-4 py-2'>
      {!chargeResult ? (
        <div className='space-y-4'>
          <div className='space-y-2'>
            <Label className='text-xs font-medium'>Metode Pembayaran Online (Tripay Payment Gateway)</Label>
            <Select value={channel} onValueChange={setChannel}>
              <SelectTrigger className='text-xs h-9'>
                <SelectValue placeholder='Pilih saluran pembayaran' />
              </SelectTrigger>
              <SelectContent>
                {CHANNELS.map((ch) => (
                  <SelectItem key={ch.value} value={ch.value} className='text-xs'>
                    {ch.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <p className='text-[11px] text-muted-foreground'>
              Biaya admin gateway akan diproses otomatis oleh Tripay.
            </p>
          </div>

          <Button
            type='button'
            className='w-full gap-2'
            onClick={handleGenerate}
            disabled={chargeMutation.isPending || outstanding <= 0}
          >
            <CreditCard className='h-4 w-4' />
            {chargeMutation.isPending
              ? 'Membuat Tagihan Online...'
              : `Generate Pembayaran Online ${formatCurrency(outstanding)}`}
          </Button>
        </div>
      ) : (
        <div className='space-y-4'>
          <Card className='border-emerald-500/30 bg-emerald-500/5'>
            <CardContent className='p-4 space-y-3'>
              <div className='flex items-center justify-between'>
                <Badge variant='outline' className='bg-emerald-500/10 text-emerald-600 border-emerald-500/30 text-xs font-semibold'>
                  MENUNGGU PEMBAYARAN
                </Badge>
                <span className='font-mono font-bold text-sm text-foreground'>
                  {formatCurrency(chargeResult.amount || outstanding)}
                </span>
              </div>

              {chargeResult.va_number && (
                <div className='rounded-lg border bg-background p-3 flex items-center justify-between'>
                  <div>
                    <p className='text-xs text-muted-foreground font-medium'>Nomor Virtual Account ({channel}):</p>
                    <p className='font-mono text-base font-bold text-foreground mt-0.5 tracking-wider'>
                      {chargeResult.va_number}
                    </p>
                  </div>
                  <Button
                    variant='outline'
                    size='sm'
                    className='gap-1 h-8 text-xs'
                    onClick={() => handleCopy(chargeResult.va_number, 'Nomor VA')}
                  >
                    {copied ? <Check className='h-3.5 w-3.5 text-emerald-600' /> : <Copy className='h-3.5 w-3.5' />}
                    Salin
                  </Button>
                </div>
              )}

              {chargeResult.payment_url && (
                <div className='flex items-center gap-2 pt-1'>
                  <Button
                    variant='default'
                    size='sm'
                    className='w-full gap-1.5'
                    onClick={() => window.open(chargeResult.payment_url, '_blank')}
                  >
                    <ExternalLink className='h-4 w-4' />
                    Buka Halaman Pembayaran / Scan QRIS
                  </Button>
                  <Button
                    variant='outline'
                    size='sm'
                    className='shrink-0'
                    onClick={() => handleCopy(chargeResult.payment_url, 'Link pembayaran')}
                  >
                    <Copy className='h-3.5 w-3.5' />
                  </Button>
                </div>
              )}
            </CardContent>
          </Card>

          <Alert className='bg-blue-500/10 border-blue-500/20 text-xs py-3'>
            <Sparkles className='h-4 w-4 text-blue-600 dark:text-blue-400 shrink-0' />
            <AlertTitle className='font-semibold text-blue-900 dark:text-blue-300'>
              Sinkronisasi Otomatis Router & Faktur
            </AlertTitle>
            <AlertDescription className='text-blue-700 dark:text-blue-400 mt-1 leading-relaxed'>
              Saat pembayaran terverifikasi oleh Tripay, invoice otomatis ditandai lunas dan pelanggan yang terisolir akan langsung dipulihkan di MikroTik tanpa konfirmasi manual.
            </AlertDescription>
          </Alert>

          <Button
            variant='ghost'
            size='sm'
            className='w-full text-xs text-muted-foreground'
            onClick={() => setChargeResult(null)}
          >
            Pilih metode / saluran pembayaran lain
          </Button>
        </div>
      )}
    </div>
  )
}
