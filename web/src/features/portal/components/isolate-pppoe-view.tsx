import { useState } from 'react'
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
} from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { toast } from 'sonner'

export function IsolatePPPoEView() {
  const [identifier, setIdentifier] = useState('')
  const [isSearching, setIsSearching] = useState(false)
  const [searched, setSearched] = useState(false)

  // Contoh data tagihan mockup saat user mencari
  const mockBill = {
    customerName: 'Pelanggan Internet Polyglot',
    customerCode: identifier ? identifier.toUpperCase() : 'CUST-00921',
    planName: 'Home Fiber Ultra 50 Mbps',
    period: 'Periode Tagihan Bulan Berjalan',
    dueDate: '20 Agustus 2026',
    amount: 175000,
    status: 'UNPAID / TERISOLIR',
  }

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    if (!identifier.trim()) {
      toast.error('Masukkan Nomor HP, Kode Pelanggan, atau Username PPPoE Anda')
      return
    }
    setIsSearching(true)
    setTimeout(() => {
      setIsSearching(false)
      setSearched(true)
      toast.success('Data tagihan ditemukan')
    }, 600)
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
            {searched && (
              <div className='rounded-xl border bg-muted/30 p-4 space-y-4 animate-in fade-in-50 duration-300'>
                <div className='flex items-start justify-between gap-2 border-b pb-3'>
                  <div>
                    <h3 className='font-semibold text-sm text-foreground'>{mockBill.customerName}</h3>
                    <p className='text-xs font-mono text-muted-foreground'>{mockBill.customerCode} • {mockBill.planName}</p>
                  </div>
                  <Badge variant='destructive' className='text-[10px] uppercase'>
                    {mockBill.status}
                  </Badge>
                </div>

                <div className='flex items-center justify-between text-sm py-1'>
                  <span className='text-muted-foreground text-xs'>Total Tagihan Belum Dibayar:</span>
                  <span className='text-lg font-bold text-red-600 dark:text-red-400 font-mono'>
                    {formatCurrency(mockBill.amount)}
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
                      <div className='p-4 bg-white rounded-lg border inline-block shadow-sm'>
                        <div className='w-36 h-36 bg-zinc-100 flex items-center justify-center rounded border border-dashed border-zinc-300 text-zinc-400 text-xs font-mono'>
                          [ QRIS BARCODE ]
                        </div>
                      </div>
                      <p className='text-xs text-muted-foreground'>
                        Scan QRIS di atas dengan <strong>BCA Mobile, Livin, GoPay, OVO, ShopeePay, DANA,</strong> atau aplikasi m-Banking apa saja.
                      </p>
                    </TabsContent>

                    <TabsContent value='va' className='space-y-2 pt-2 text-xs'>
                      <div className='rounded-lg border p-3 bg-background flex items-center justify-between'>
                        <div>
                          <p className='font-semibold'>BCA Virtual Account</p>
                          <p className='font-mono text-muted-foreground text-[11px]'>8277 0812 3456 7890</p>
                        </div>
                        <Button variant='outline' size='sm' onClick={() => { navigator.clipboard.writeText('8277081234567890'); toast.success('Nomor VA disalin') }}>
                          Salin
                        </Button>
                      </div>
                      <div className='rounded-lg border p-3 bg-background flex items-center justify-between'>
                        <div>
                          <p className='font-semibold'>BRI Virtual Account (BRIVA)</p>
                          <p className='font-mono text-muted-foreground text-[11px]'>1029 8812 3456 7890</p>
                        </div>
                        <Button variant='outline' size='sm' onClick={() => { navigator.clipboard.writeText('1029881234567890'); toast.success('Nomor VA disalin') }}>
                          Salin
                        </Button>
                      </div>
                      <div className='rounded-lg border p-3 bg-background flex items-center justify-between'>
                        <div>
                          <p className='font-semibold'>Mandiri Virtual Account</p>
                          <p className='font-mono text-muted-foreground text-[11px]'>8890 0812 3456 7890</p>
                        </div>
                        <Button variant='outline' size='sm' onClick={() => { navigator.clipboard.writeText('8890081234567890'); toast.success('Nomor VA disalin') }}>
                          Salin
                        </Button>
                      </div>
                    </TabsContent>

                    <TabsContent value='retail' className='space-y-2 pt-2 text-xs'>
                      <div className='rounded-lg border p-3 bg-background flex items-center justify-between'>
                        <div>
                          <p className='font-semibold'>Indomaret</p>
                          <p className='font-mono text-muted-foreground text-[11px]'>Kode Bayar: PLG-998123</p>
                        </div>
                        <Button variant='outline' size='sm' onClick={() => { navigator.clipboard.writeText('PLG-998123'); toast.success('Kode bayar disalin') }}>
                          Salin
                        </Button>
                      </div>
                      <div className='rounded-lg border p-3 bg-background flex items-center justify-between'>
                        <div>
                          <p className='font-semibold'>Alfamart / Alfamidi</p>
                          <p className='font-mono text-muted-foreground text-[11px]'>Kode Bayar: PLG-998123</p>
                        </div>
                        <Button variant='outline' size='sm' onClick={() => { navigator.clipboard.writeText('PLG-998123'); toast.success('Kode bayar disalin') }}>
                          Salin
                        </Button>
                      </div>
                    </TabsContent>
                  </Tabs>
                </div>
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
