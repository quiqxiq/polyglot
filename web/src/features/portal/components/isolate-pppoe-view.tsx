import { useState, useEffect } from 'react'
import {
  ShieldAlert,
  Phone,
  QrCode,
  CreditCard,
  Building2,
  ArrowRight,
  WifiOff,
  Sparkles,
  HelpCircle,
  Clock,
  RefreshCw,
  ExternalLink,
} from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { toast } from 'sonner'

interface PublicBill {
  customer_name: string
  customer_code: string
  invoice_id: string
  invoice_number: string
  period: string
  total: number
  paid_amount: number
  outstanding: number
  due_date: string
  status: string
  manual_payment_code: string
}

export function IsolatePPPoEView() {
  const [identifier, setIdentifier] = useState('')
  const [isSearching, setIsSearching] = useState(false)
  const [searched, setSearched] = useState(false)
  const [bill, setBill] = useState<PublicBill | null>(null)
  const [chargeResult, setChargeResult] = useState<{
    external_id?: string
    payment_url?: string
    qr_string?: string
    va_number?: string
    status?: string
    amount?: number
  } | null>(null)
  const [isGenerating, setIsGenerating] = useState(false)

  const fetchBill = async (query: string) => {
    const term = query.trim()
    if (!term) {
      toast.error('Masukkan Nomor HP, Kode Pelanggan, atau Kode Bayar Anda')
      return
    }
    setIsSearching(true)
    setChargeResult(null)
    try {
      const baseUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080'
      const res = await fetch(`${baseUrl}/api/portal/bill?identifier=${encodeURIComponent(term)}`)
      if (!res.ok) {
        setBill(null)
        setSearched(true)
        toast.error('Tagihan tidak ditemukan atau sudah lunas')
        return
      }
      const data: PublicBill = await res.json()
      setBill(data)
      setSearched(true)
      toast.success(`Tagihan untuk ${data.customer_name} ditemukan`)
    } catch (err) {
      setBill(null)
      setSearched(true)
      toast.error('Gagal mengambil data tagihan', {
        description: err instanceof Error ? err.message : undefined,
      })
    } finally {
      setIsSearching(false)
    }
  }

  // Auto-search jika parameter query tersedia di URL
  useEffect(() => {
    if (typeof window === 'undefined') return
    const params = new URLSearchParams(window.location.search)
    const paramId = params.get('identifier') || params.get('inv') || params.get('code') || params.get('user')
    if (paramId) {
      setIdentifier(paramId)
      fetchBill(paramId)
    }
  }, [])

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    fetchBill(identifier)
  }

  const handleGenerateCharge = async (channel: string) => {
    if (!bill) {
      toast.error('Pilih atau cari tagihan terlebih dahulu')
      return
    }
    setIsGenerating(true)
    try {
      const baseUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080'
      const res = await fetch(`${baseUrl}/api/portal/charge`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ invoice_id: bill.invoice_id, channel, expire_minutes: 60 }),
      })
      if (!res.ok) {
        const data = await res.json().catch(() => null)
        throw new Error(data?.message || 'Gagal memproses pembayaran online')
      }
      const data = await res.json()
      setChargeResult(data)
      toast.success('Tagihan online Tripay berhasil dibuat')
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Gagal membuat tagihan online'
      toast.error('Gagal membuat tagihan online', { description: message })
    } finally {
      setIsGenerating(false)
    }
  }

  const formatCurrency = (val: number) => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      maximumFractionDigits: 0,
    }).format(val)
  }

  return (
    <div className='min-h-screen bg-gradient-to-b from-background via-muted/30 to-background flex flex-col items-center justify-center p-4 sm:p-6 lg:p-8'>
      {/* Container utama */}
      <div className='w-full max-w-2xl space-y-6'>
        {/* Header Branding & Alert */}
        <div className='text-center space-y-3'>
          <div className='inline-flex items-center justify-center p-3 rounded-2xl bg-destructive/10 text-destructive border border-destructive/20 shadow-sm animate-pulse'>
            <WifiOff className='h-8 w-8' />
          </div>
          <div className='space-y-1'>
            <Badge variant='outline' className='bg-red-500/10 text-red-600 dark:text-red-400 border-red-500/30 text-xs px-3 py-1 font-semibold gap-1.5'>
              <ShieldAlert className='h-3.5 w-3.5' /> STATUS LAYANAN: TERISOLIR SEMENTARA
            </Badge>
            <h1 className='text-2xl sm:text-3xl font-bold tracking-tight text-foreground'>
              Pemberitahuan Penangguhan Internet PPPoE
            </h1>
            <p className='text-sm text-muted-foreground max-w-lg mx-auto'>
              Koneksi internet rumah Anda sementara dialihkan ke halaman ini karena terdapat tagihan bulanan yang melewati tanggal jatuh tempo.
            </p>
          </div>
        </div>

        {/* Card Cek & Bayar Tagihan */}
        <Card className='border-border/80 shadow-lg bg-card/95 backdrop-blur-sm'>
          <CardHeader className='pb-4 border-b'>
            <CardTitle className='text-base sm:text-lg flex items-center justify-between'>
              <span>Cek Tagihan & Buka Isolir Otomatis</span>
              <span className='text-xs font-normal text-muted-foreground flex items-center gap-1'>
                <Clock className='h-3.5 w-3.5 text-amber-500' /> Buka Otomatis 24 Jam
              </span>
            </CardTitle>
            <CardDescription className='text-xs'>
              Masukkan Nomor HP, Kode Pelanggan, atau Username PPPoE yang terdaftar untuk melihat tagihan dan melakukan pembayaran instan.
            </CardDescription>
          </CardHeader>

          <CardContent className='pt-5 space-y-5'>
            {/* Form Pencarian */}
            <form onSubmit={handleSearch} className='flex gap-2'>
              <div className='relative flex-1'>
                <Input
                  value={identifier}
                  onChange={(e) => setIdentifier(e.target.value)}
                  placeholder='Contoh: 081234567890 atau CUST-00123'
                  className='h-10 text-sm font-mono'
                />
              </div>
              <Button type='submit' disabled={isSearching} className='h-10 gap-1.5 bg-primary text-primary-foreground'>
                {isSearching ? <RefreshCw className='h-4 w-4 animate-spin' /> : <ArrowRight className='h-4 w-4' />}
                Cek Tagihan
              </Button>
            </form>

            {/* Rincian Tagihan */}
            {searched && bill && (
              <div className='rounded-xl border bg-muted/30 p-4 space-y-4 animate-in fade-in-50 duration-300'>
                <div className='flex items-start justify-between gap-2 border-b pb-3'>
                  <div>
                    <h3 className='font-semibold text-sm text-foreground'>{bill.customer_name}</h3>
                    <p className='text-xs font-mono text-muted-foreground'>
                      {bill.customer_code} • No. Faktur {bill.invoice_number} ({bill.period})
                    </p>
                  </div>
                  <Badge variant={chargeResult?.status === 'PAID' ? 'default' : 'destructive'} className='text-[10px] uppercase'>
                    {chargeResult?.status === 'PAID'
                      ? 'LUNAS / PULIH'
                      : bill.status === 'OVERDUE'
                      ? 'JATUH TEMPO'
                      : 'TERISOLIR'}
                  </Badge>
                </div>

                <div className='flex items-center justify-between text-sm py-1'>
                  <span className='text-muted-foreground text-xs'>Total Tagihan Belum Dibayar:</span>
                  <span className='text-lg font-bold text-red-600 dark:text-red-400 font-mono'>
                    {formatCurrency(chargeResult?.amount || bill.outstanding)}
                  </span>
                </div>

                {/* Metode Pembayaran */}
                <div className='space-y-3 pt-2'>
                  <Label className='text-xs font-semibold'>Pilih Metode Pembayaran Cepat:</Label>
                  <Tabs defaultValue='qris' className='w-full'>
                    <TabsList className='grid grid-cols-3 w-full h-9'>
                      <TabsTrigger value='qris' className='text-xs gap-1.5'>
                        <QrCode className='h-3.5 w-3.5' /> QRIS
                      </TabsTrigger>
                      <TabsTrigger value='va' className='text-xs gap-1.5'>
                        <CreditCard className='h-3.5 w-3.5' /> Virtual Account
                      </TabsTrigger>
                      <TabsTrigger value='retail' className='text-xs gap-1.5'>
                        <Building2 className='h-3.5 w-3.5' /> Gerai Retail
                      </TabsTrigger>
                    </TabsList>

                    <TabsContent value='qris' className='space-y-3 pt-3 text-center'>
                      {chargeResult?.payment_url ? (
                        <div className='space-y-3'>
                          <div className='p-4 bg-emerald-500/10 border border-emerald-500/20 rounded-lg text-center space-y-2'>
                            <p className='text-xs font-semibold text-emerald-800 dark:text-emerald-300'>Tagihan Online Tripay Siap Dibayar</p>
                            <Button
                              className='w-full gap-2'
                              onClick={() => window.open(chargeResult.payment_url, '_blank')}
                            >
                              <ExternalLink className='h-4 w-4' />
                              Buka Halaman Scan QRIS Sekarang
                            </Button>
                          </div>
                        </div>
                      ) : (
                        <div className='space-y-3'>
                          <Button
                            className='w-full gap-2'
                            onClick={() => handleGenerateCharge('QRIS')}
                            disabled={isGenerating}
                          >
                            <QrCode className='h-4 w-4' />
                            {isGenerating ? 'Menghubungkan ke Tripay...' : 'Buat Kode QRIS Resmi Tripay'}
                          </Button>
                          <p className='text-xs text-muted-foreground'>
                            Mendukung <strong>BCA Mobile, Livin by Mandiri, GoPay, OVO, ShopeePay, DANA,</strong> atau aplikasi m-Banking apa saja.
                          </p>
                        </div>
                      )}
                    </TabsContent>

                    <TabsContent value='va' className='space-y-2 pt-2 text-xs'>
                      {chargeResult?.va_number ? (
                        <div className='rounded-lg border bg-emerald-500/10 border-emerald-500/30 p-3 flex items-center justify-between'>
                          <div>
                            <p className='font-semibold text-emerald-800 dark:text-emerald-300'>Nomor Virtual Account Anda</p>
                            <p className='font-mono font-bold text-sm mt-0.5'>{chargeResult.va_number}</p>
                          </div>
                          <Button
                            variant='outline'
                            size='sm'
                            onClick={() => {
                              navigator.clipboard.writeText(chargeResult.va_number || '')
                              toast.success('Nomor VA disalin')
                            }}
                          >
                            Salin
                          </Button>
                        </div>
                      ) : (
                        <div className='space-y-2'>
                          <Button
                            variant='outline'
                            className='w-full justify-between h-10 text-xs'
                            onClick={() => handleGenerateCharge('BRIVA')}
                            disabled={isGenerating}
                          >
                            <span>BRI Virtual Account (BRIVA)</span>
                            <span className='text-[11px] text-muted-foreground'>Klik untuk buat VA</span>
                          </Button>
                          <Button
                            variant='outline'
                            className='w-full justify-between h-10 text-xs'
                            onClick={() => handleGenerateCharge('BNIVA')}
                            disabled={isGenerating}
                          >
                            <span>BNI Virtual Account (BNIVA)</span>
                            <span className='text-[11px] text-muted-foreground'>Klik untuk buat VA</span>
                          </Button>
                          <Button
                            variant='outline'
                            className='w-full justify-between h-10 text-xs'
                            onClick={() => handleGenerateCharge('MANDIRIVA')}
                            disabled={isGenerating}
                          >
                            <span>Mandiri Virtual Account</span>
                            <span className='text-[11px] text-muted-foreground'>Klik untuk buat VA</span>
                          </Button>
                        </div>
                      )}
                    </TabsContent>

                    <TabsContent value='retail' className='space-y-2 pt-2 text-xs'>
                      <div className='rounded-lg border p-3 bg-background flex items-center justify-between'>
                        <div>
                          <p className='font-semibold'>Indomaret</p>
                          <p className='font-mono text-muted-foreground text-[11px]'>
                            Kode Bayar: {bill.manual_payment_code || bill.invoice_number}
                          </p>
                        </div>
                        <Button
                          variant='outline'
                          size='sm'
                          onClick={() => {
                            navigator.clipboard.writeText(bill.manual_payment_code || bill.invoice_number)
                            toast.success('Kode bayar disalin')
                          }}
                        >
                          Salin
                        </Button>
                      </div>
                      <div className='rounded-lg border p-3 bg-background flex items-center justify-between'>
                        <div>
                          <p className='font-semibold'>Alfamart / Alfamidi</p>
                          <p className='font-mono text-muted-foreground text-[11px]'>
                            Kode Bayar: {bill.manual_payment_code || bill.invoice_number}
                          </p>
                        </div>
                        <Button
                          variant='outline'
                          size='sm'
                          onClick={() => {
                            navigator.clipboard.writeText(bill.manual_payment_code || bill.invoice_number)
                            toast.success('Kode bayar disalin')
                          }}
                        >
                          Salin
                        </Button>
                      </div>
                    </TabsContent>
                  </Tabs>
                </div>
              </div>
            )}

            {/* State Tagihan Tidak Ditemukan */}
            {searched && !bill && !isSearching && (
              <div className='rounded-xl border border-destructive/20 bg-destructive/5 p-4 text-center space-y-2 animate-in fade-in-50'>
                <p className='text-sm font-semibold text-destructive'>Tagihan Tidak Ditemukan</p>
                <p className='text-xs text-muted-foreground max-w-md mx-auto'>
                  Tidak ada tagihan tertunggak yang ditemukan untuk data tersebut. Jika status internet Anda masih terisolir, mohon hubungi layanan pelanggan kami.
                </p>
              </div>
            )}

            {/* Jaminan Auto-Restore */}
            <Alert className='bg-emerald-500/10 border-emerald-500/20 text-xs py-3'>
              <Sparkles className='h-4 w-4 text-emerald-600 dark:text-emerald-400 shrink-0' />
              <AlertTitle className='font-semibold text-emerald-900 dark:text-emerald-300'>
                Pembukaan Isolir Otomatis (Instant Restore)
              </AlertTitle>
              <AlertDescription className='text-emerald-700 dark:text-emerald-400 mt-1 leading-relaxed'>
                Begitu pembayaran Anda terverifikasi oleh sistem payment gateway, profil internet Anda di router akan dipulihkan secara otomatis dalam <strong>5 - 10 detik</strong> tanpa perlu konfirmasi manual atau me-restart router.
              </AlertDescription>
            </Alert>
          </CardContent>

          <CardFooter className='border-t pt-4 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs text-muted-foreground'>
            <div className='flex items-center gap-1.5'>
              <HelpCircle className='h-4 w-4 text-blue-500' />
              <span>Butuh bantuan atau konfirmasi transfer manual?</span>
            </div>
            <Button
              variant='outline'
              size='sm'
              className='gap-1.5 text-emerald-600 hover:text-emerald-700 border-emerald-500/30 hover:bg-emerald-500/10 w-full sm:w-auto'
              onClick={() => window.open('https://wa.me/', '_blank')}
            >
              <Phone className='h-3.5 w-3.5' /> Hubungi CS via WhatsApp
            </Button>
          </CardFooter>
        </Card>
      </div>
    </div>
  )
}
